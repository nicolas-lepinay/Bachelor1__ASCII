package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

// ⭐ DÉCLARATION DES VARIABLES GLOBALES :
var myTemplates *template.Template

var fonts struct {
	Standard   []string
	Shadow     []string
	Thinkertoy []string
}
var Send struct {
	Input string
	Font  string
	Art   string
}

// ⭐⭐⭐ FONCTION MAIN ⭐⭐⭐ \\
func main() {

	// Charger les fichiers du dossier 'static' (style.css, etc.) sur le serveur, pour pouvoir les utiliser :
	fs := http.FileServer(http.Dir("./static/"))              // ⚠️ Ne pas oublier (1) le point, et (2) le slash à la fin : sinon, ça ne fonctionne pas, et les fichiers présents dans le dossier 'static' ne seront pas chargés !
	http.Handle("/static/", http.StripPrefix("/static/", fs)) // ⚠️ Ne pas oublier le slash à la fin, sinon cela ne fonctionne pas.

	// Lecture du template :
	myTemplates = template.Must(template.ParseGlob("./static/index.html")) // .Must() : renvoie une erreur si le template ne peut pas être chargé. | .ParseGlob() : permet de charger plusieurs templates, si nécessaire.

	// Lectures des polices de caractères :
	fonts.Standard = readFont("standard")
	fonts.Shadow = readFont("shadow")
	fonts.Thinkertoy = readFont("thinkertoy")

	// Fonction HandleFunc :
	http.HandleFunc("/", indexHandler)

	// Lancement du serveur :
	fmt.Println("Listening server at port 8080.")
	http.ListenAndServe(":8080", nil)
}

// ⭐ FONCTION 'INDEXHANDLER' POUR LE HANDLEFUNC :
func indexHandler(w http.ResponseWriter, r *http.Request) {

	// GESTION DU STATUT '404' :
	if r.URL.Path != "/" {
		http.Error(w, "404 PAGE NOT FOUND", http.StatusNotFound)
		return
	}

	// GESTION DES REQUEST METHODS :
	switch r.Method {

	// 🍔 Méthode 'GET' — Lorsqu'on charge la page pour la première fois :
	case "GET":
		fmt.Println("Method is: ", r.Method)

		Send.Input = ""
		Send.Font = "standard"
		Send.Art = generator("* Bienvenue *", "standard") // Message par défaut.

		myTemplates.ExecuteTemplate(w, "index.html", Send) // Seule ligne indispensable dans le cas où la méthode est 'GET', car il faut quand même charger le template pour ne pas avoir une page blanche.

	// 🍔 Méthode 'POST' — Lorsqu'on appuie sur le bouton 'Create' ou 'Download' pour générer de l'ASCII Art :
	case "POST":
		fmt.Println("Method is: ", r.Method)

		var input string
		var font string      // 'Standard', 'Shadow' ou 'Thinkertoy'
		var genOrDown string // 'Generate' ou 'Download'

		body, _ := ioutil.ReadAll(r.Body)        // body est un tableau d'uint8 : [116 101 120 116 84 111 80 ...], mais chaque unint8 représente un caractère ASCII. Une fois décodé, body = [textToPrint=Hello+World&font=standard&genOrDown=generate]. Donc body contient toutes les valeurs des paramètres existants dans le template HTML.
		query, _ := url.ParseQuery(string(body)) // .ParseQuery() analyse body (casté en string) et crée une map en fonction des caractères '&' et '='. query est donc une map contenant tous les paramètres trouvés et leur valeur : ici, ce sont les sélecteurs "font", "genOrDown" et l'input "textToPrint" dans le template HTML. Chaque valeur est elle-même contenue dans un array : query = map[font: [standard] genOrDown: [generate] textToPrint: [Hello World]]

		// Stockage des valeurs demandées dans les variables input, font et genOrDown :
		for key, value := range query {
			switch key {
			case "textToPrint":
				input = value[0]
			case "font":
				font = value[0]
			case "genOrDown":
				genOrDown = value[0]
				fmt.Println("Button clicked: ", genOrDown)
			default:
				http.Error(w, "400 Bad Request", 400)
				return
			}
		}

		// Génération de l'ASCII Art :
		ascii_art := generator(input, font)
		switch genOrDown {
		case "generate": // On écrit l'ASCII Art dans le template :
			Send.Input = input
			Send.Font = font
			Send.Art = ascii_art

			// Écriture dans le template 'index.html' :
			myTemplates.ExecuteTemplate(w, "index.html", Send)

		case "download": // Serving file to download :
			file := strings.NewReader(ascii_art)
			fileSize := strconv.FormatInt(file.Size(), 10)
			w.Header().Set("Content-Disposition", "attachment; filename=ASCII_Art.txt")
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Content-Length", fileSize)
			file.Seek(0, 0)
			io.Copy(w, file)
		}
	}
}

// ⭐ FONCTION DE LECTURE DES POLICES :
func readFont(font string) []string {
	var lines []string

	file, _ := os.Open("static/" + font + ".txt") // Ouverture du fichier.
	defer file.Close()

	scanner := bufio.NewScanner(file) // Création d'un scanner à partir du fichier.
	for scanner.Scan() {
		lines = append(lines, scanner.Text()) // J'ajoute 'scanner' (transformé en texte) à l'array 'lines'.
	}
	return lines
}

// ⭐ FONCTION GÉNÉRATRICE D'ASCII ART :
func generator(input, font string) string {
	var lines []string
	var res string

	switch font {
	case "standard":
		lines = fonts.Standard
	case "shadow":
		lines = fonts.Shadow
	case "thinkertoy":
		lines = fonts.Thinkertoy
	}

	words := strings.Split(input, "\\n")
	for _, word := range words {
		for i := 0; i < 8; i++ {
			for _, char := range word {
				if char > 31 && char < 127 {
					res = res + lines[(int(char)-32)*8+i]
				}
			}
			res += "\n"
		}
	}
	return res
}
