package main
 
import (
	"fmt"
	"strings"
	"net"
	"os"
	"runtime"
	"io/ioutil"
	"time"
	"sync"
	"log"
	"net/http"
	"strconv"
	"github.com/gorilla/mux"
)

type SafeList struct {
	fournisseur []string
	mux sync.Mutex
}

var broadcast string = "255.255.255.255"
var port string = "8080"
var fournisseurs SafeList
var key string


//Gestion du timeout lors des requêtes à destination des fournisseurs
var timeout = time.Duration(10 *time.Second)
var client  = http.Client{
	Timeout: timeout,
}


func updateSlice(f string) (bool) {
fournisseurs.mux.Lock()
	for i := 0; i < len(fournisseurs.fournisseur); i++ {
		if fournisseurs.fournisseur[i] == f {

			fournisseurs.mux.Unlock()
			return false;
		}
	}

	fournisseurs.fournisseur = append(fournisseurs.fournisseur, f)
fournisseurs.mux.Unlock()
	return true
}

func parseRequest(object string) ([]string) {
	msg := strings.Fields(object)

	return msg
}


func findFournisseur(f string) (bool) {
fournisseurs.mux.Lock()
	for i:= 0; i < len(fournisseurs.fournisseur); i++ {
		if f == fournisseurs.fournisseur[i] {

			fournisseurs.mux.Unlock()
			return true
		}
	}

fournisseurs.mux.Unlock()
	return false;
}

func reset(){
fournisseurs.mux.Lock()
	
fournisseurs.fournisseur = nil
	
fournisseurs.mux.Unlock()
}

func CheckError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
        os.Exit(0)
    }
}


func sendBroadcast() {
for{
    reset()
    ServerAddr,err := net.ResolveUDPAddr("udp","255.255.255.255:8081")
    CheckError(err)
 
    LocalAddr, err := net.ResolveUDPAddr("udp", ":0")
    CheckError(err)
 
    Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
    CheckError(err)
 
 s := []string{key, port};
buf := []byte(strings.Join(s, " "))
    defer Conn.Close()
    for {
        _,err := Conn.Write(buf)
        if err != nil {
            fmt.Println(key, err)
        }
        time.Sleep(time.Second * 30)
    }



}

}



func help(w http.ResponseWriter, r *http.Request) {
   w.WriteHeader(http.StatusOK)
  fmt.Fprintf(w, "\n\nMenu aide - Fonctionnalités disponibles\n1. \"/doProductsList\" => Pour lister l'ensemble des produits\n2. \"/doBuyProduct/produit/quantité\" => Pour acheter une quantité d'un produit \n\n")
}

func doProductsList(w http.ResponseWriter, r *http.Request) {

	listeProduits:= make([]string,0)	

	//Recuperation des fournisseurs
	fournisseurs.mux.Lock()
	
	for _,fournisseur := range fournisseurs.fournisseur {
		resp,err:= client.Get("http://"+fournisseur+"/myfile/stock/")

		//Si tout s'est bien passé et qu'il n'y a pas d'erreur
    	if err == nil {
    				
    		defer resp.Body.Close()
			body,_ := ioutil.ReadAll(resp.Body)
			listeProduits = append(listeProduits,string(body))
    	}

	}
	fournisseurs.mux.Unlock()
    w.WriteHeader(http.StatusOK)
  fmt.Fprintf(w, "Liste des produits :\n%s\n",listeProduits)
						
}

func APIrest (){

router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/doProductsList", doProductsList).Methods("GET")
	router.HandleFunc("/doBuyProduct/{x}/{y}", doBuyProduct).Methods("GET")
	router.HandleFunc("/", help).Methods("GET")

	log.Fatal(http.ListenAndServe(":8082", router))


}
func doBuyProduct(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
    var response bool = false

	product := vars["x"]
	occ,_ := strconv.Atoi(vars["y"])

	//Recuperation des fournisseurs
	fournisseurs.mux.Lock()

	for _,fournisseur := range fournisseurs.fournisseur {
		
		resp,err:= client.Get("http://"+fournisseur+"/myfile/dispo/"+product)
		
		//Si tout s'est bien passé et qu'il n'y a pas d'erreur
		if err == nil {

			defer resp.Body.Close()
			body,_ := ioutil.ReadAll(resp.Body)

		
    		quantity,_:=strconv.Atoi(string(body)) 
    		
    		if occ <= quantity {

    			//On essaye d'effectuer l'achat du produit pour le client
    			resp2,err2:= client.Get("http://"+fournisseur+"/myfile/buy/"+product+"/"+strconv.Itoa(occ))
    			
    			//Si tout s'est bien passé et qu'il n'y a pas d'erreur
    			if err2 == nil {
    				defer resp2.Body.Close()
					body2,_ := ioutil.ReadAll(resp2.Body)
					//Si l'achat s'est bien effectue
					if strings.Compare(string(body2),"OK")==0 {
     					w.WriteHeader(http.StatusOK)
						fmt.Fprintf(w, "\nVous avez acheté %d  %s(s)\n",occ,product)
						fournisseurs.mux.Unlock()
						return
    				}
    			
    			}
    		
    		}
		}
	}
	fournisseurs.mux.Unlock()
	if !response {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Produit non disponible! \n")
	}
}



func main() {
	runtime.GOMAXPROCS(3)
	b, err := ioutil.ReadFile("key.txt") // just pass the file name
   	 if err != nil {
      	  fmt.Print(err)
    	}

   	 key = string(b) // convert content to a 'string'
	 key = key[0:len(key)-1]
	go sendBroadcast()	



ServerAddr,err := net.ResolveUDPAddr("udp",":"+port)
	CheckError(err)
    
	//Now listen at selected port 
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()
 
	buf := make([]byte, 1024)
/*
	//API REST
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/doProductsList", doProductsList).Methods("GET")
	router.HandleFunc("/doBuyProduct/{x}/{y}", doBuyProduct).Methods("GET")
	router.HandleFunc("/", help).Methods("GET")

	log.Fatal(http.ListenAndServe(":8082", router))

*/
go APIrest()
for {
	
        n,addr,err := ServerConn.ReadFromUDP(buf)

	ip := strings.Split(addr.String(), ":")

 	msg := parseRequest(string(buf[0:n]))

	updateSlice(ip[0]+":"+msg[1])
	


        if err != nil {
            fmt.Println("Error: ",err)
        } 

    }
}




 
