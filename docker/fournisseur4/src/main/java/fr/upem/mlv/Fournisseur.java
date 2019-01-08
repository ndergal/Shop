package fr.upem.mlv;

import java.io.BufferedReader;
import java.io.IOException;
import java.io.InputStreamReader;
import java.io.PrintWriter;
import java.net.DatagramPacket;
import java.net.DatagramSocket;
import java.net.InetAddress;
import java.net.InetSocketAddress;
import java.net.ServerSocket;
import java.net.Socket;
import java.net.URI;
import java.nio.file.Paths;
import java.util.AbstractMap.SimpleImmutableEntry;
import java.util.HashMap;
import java.util.HashSet;
import java.util.Map;
import java.util.Map.Entry;
import java.util.Scanner;
import java.util.Set;

import javax.ws.rs.GET;
import javax.ws.rs.Path;
import javax.ws.rs.PathParam;
import javax.ws.rs.core.Context;

import org.glassfish.grizzly.http.server.HttpServer;
import org.glassfish.jersey.grizzly2.httpserver.GrizzlyHttpServerFactory;
import org.glassfish.jersey.server.ResourceConfig;

/**
 * Main class.
 *
 */
@Path("myfile")
public class Fournisseur {

	public static final Stock<String, Integer> stock;
	public static final Clients<String> client;
	public static final String BASE_URI;

	static {
		stock = new Stock<>();
		client = new Clients<>();
		BASE_URI = "http://0.0.0.0:8080/";
	}

	static class Clients<R> {
		private final Set<R> clients = new HashSet<>();
		private final Object monitor = new Object();

		public void add(R client) {
			synchronized (monitor) {
				clients.add(client);
			}
		}

		public boolean contains(R client) {
			synchronized (monitor) {
				return clients.contains(client);
			}
		}

		public boolean remove(R client) {
			synchronized (monitor) {
				return clients.remove(client);
			}
		}
	}

	static class Stock<R, V> {
		private final Map<R, V> stock = new HashMap<>();
		private final Object monitor = new Object();

		public Set<Entry<R, V>> entrySet() {
			synchronized (monitor) {

				Set<Entry<R, V>> set = new HashSet<>();
				Set<Entry<R, V>> setReal = stock.entrySet();
				setReal.forEach(e -> {
					set.add(new SimpleImmutableEntry<R, V>(e.getKey(), e.getValue()));
				});
				return set;
			}
		}

		public void put(R key, V value) {
			synchronized (monitor) {
				stock.put(key, value);
			}
		}

		public V get(R key) {
			synchronized (monitor) {
				return stock.get(key);
			}
		}
	}

	public static HttpServer startServer() throws IOException {
		try (Scanner sc = new Scanner(Paths.get("/opt/app/stock.txt"))) {
			while (sc.hasNextLine()) {
				String[] str = sc.nextLine().split("\t");
				stock.put(str[0], Integer.parseInt(str[1]));
			}
		}
		ResourceConfig rc = new ResourceConfig().packages("fr.upem.mlv");
		return GrizzlyHttpServerFactory.createHttpServer(URI.create(BASE_URI), rc);
	}

	@Path("stock")
	@GET
	public String stock(@Context org.glassfish.grizzly.http.server.Request re) throws IOException {
		if (!client.contains(re.getRemoteAddr())) {
			return "Get OUT !!";
		}

		StringBuilder sb = new StringBuilder();

		stock.entrySet().forEach(e -> {
			sb.append(e.getKey() + " : " + e.getValue() + "\n");
		});

		return sb.toString();

	}

	@Path("dispo/{id}")
	@GET
	public String article(@PathParam("id") String id, @Context org.glassfish.grizzly.http.server.Request re)
			throws IOException {
		if (!client.contains(re.getRemoteAddr())) {
			return "Get OUT !!";
		}
		Integer value;

		value = stock.get(id);

		if (value == null) {
			return Integer.toString(0);
		}
		return Integer.toString(value);
	}

	@Path("buy/{id}/{nb}")
	@GET
	public String buy(@PathParam("id") String id, @PathParam("nb") int nb,
			@Context org.glassfish.grizzly.http.server.Request re) throws IOException {
		if (!client.contains(re.getRemoteAddr())) {
			return "Get OUT !!";
		}
		Integer value;

		value = stock.get(id);

		if (value == null || value == 0 || value < nb) {
			return "NOT OK";
		}
		stock.put(id, value - nb);
		return "OK";
	}

	public static void main(String[] args) throws IOException {
		HttpServer server = startServer();
		try (Scanner sc = new Scanner(Paths.get("/opt/app/key.txt"))) {
			while (sc.hasNextLine()) {
				MyServer srv = new MyServer(sc.nextLine());
				srv.serve();
			}
		}

	}

	static class MyServer {
		private final DatagramSocket serverSocket;
		private final String key;


		public MyServer(String key) throws IOException {
			serverSocket = new DatagramSocket(8081);
			this.key = key;
		}

		public void serve() throws IOException {
			while (!Thread.interrupted()) {


				byte[] buf = new byte[256];
				DatagramPacket packet = new DatagramPacket(buf, buf.length);
				serverSocket.receive(packet);
				InetAddress address = packet.getAddress();
				int port = packet.getPort();
				String[] received = new String(packet.getData(), 0, packet.getLength()).split(" ");
				String receivedkey = received[0];
				


				if (receivedkey.equals(key)) {

					client.add(address.getHostAddress());
					
					byte[] sendData = ("fournisseur " + Integer.toString(8080)).getBytes();
					DatagramPacket sendPacket = new DatagramPacket(sendData, sendData.length, address,
							Integer.parseInt(received[1]));
					serverSocket.send(sendPacket);

					
				}
				
			}
			serverSocket.close();
		}

	}
}
