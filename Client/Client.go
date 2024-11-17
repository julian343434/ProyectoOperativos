package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	// Verificar argumentos
	if len(os.Args) != 4 {
		fmt.Println("Uso: go run Client.go <IP> <Puerto> <Periodo>")
		return
	}

	ip := os.Args[1]
	puerto := os.Args[2]
	periodo := os.Args[3]

	// Conectar al servidor
	conn, err := net.Dial("tcp", ip+":"+puerto)
	if err != nil {
		fmt.Println("Error conectando al servidor:", err)
		return
	}
	defer conn.Close()

	fmt.Printf("Conectando al servidor en IP: %s y puerto: %s\n", ip, puerto)
	fmt.Printf("Periodo de reporte: %s\n", periodo)

	// Leer respuesta del servidor
	reader := bufio.NewReader(conn)
	resp, _ := reader.ReadString('\n')
	fmt.Print("Respuesta del servidor: " + resp)

	// Solicitar las credenciales al usuario
	fmt.Print("Ingrese su usuario: ")
	username := leerEntrada()
	fmt.Print("Ingrese su contraseña: ")
	password := leerEntrada()

	// Enviar las credenciales al servidor en formato "usuario:contraseña"
	credenciales := fmt.Sprintf("%s:%s", username, password)
	conn.Write([]byte(credenciales + "\n"))

	// Leer la respuesta del servidor después de enviar las credenciales
	resp, _ = reader.ReadString('\n')
	fmt.Println("Respuesta del servidor: " + resp)
}

// Función auxiliar para leer entrada del usuario
func leerEntrada() string {
	reader := bufio.NewReader(os.Stdin)
	entrada, _ := reader.ReadString('\n')
	return strings.TrimSpace(entrada)
}
