package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	fmt.Print("Ingrese IP: ")
	ip := leerEntrada()

	fmt.Print("Ingrese puerto: ")
	puerto := leerEntrada()

	serverIP := ip + ":" + puerto

	tcpAddress, err := net.ResolveTCPAddr("tcp4", serverIP)
	if err != nil {
		fmt.Println("Error resolviendo dirección:", err)
		return
	}

	conn, err := net.DialTCP("tcp", nil, tcpAddress)
	if err != nil {
		fmt.Println("Error conectando al servidor:", err)
		return
	}
	defer conn.Close()

	fmt.Println("Conectado al servidor:", conn.RemoteAddr())

	// Autenticación
	if !autenticar(conn) {
		fmt.Println("Autenticación fallida. Cerrando conexión.")
		return
	}

	// Bucle principal para enviar comandos al servidor
	for {
		fmt.Print("\nIngrese el comando a enviar (o 'exit' para salir): ")
		comando := leerEntrada()
		fmt.Println("Procesando comando...")

		conn.Write([]byte(comando + "\n"))

		if strings.TrimSpace(comando) == "exit" {
			fmt.Println("Cerrando conexión...")
			break
		}

		recibirRespuesta(conn)
	}
}

func autenticar(conn net.Conn) bool {
	fmt.Print("Usuario: ")
	usuario := leerEntrada()
	fmt.Print("Contraseña: ")
	contrasena := leerEntrada()

	authData := usuario + ":" + contrasena
	conn.Write([]byte(authData + "\n"))

	// Leer respuesta de autenticación
	reader := bufio.NewReader(conn)
	respuesta, _ := reader.ReadString('\n')
	return strings.TrimSpace(respuesta) == "OK"
}

func recibirRespuesta(conn net.Conn) {
	reader := bufio.NewReader(conn)
	fmt.Println("Esperando respuesta del servidor...")

	for {
		linea, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error recibiendo datos:", err)
			break
		}
		linea = strings.TrimSpace(linea)

		if strings.HasPrefix(linea, "Estado de la máquina:") {
			fmt.Println("\n" + linea) // Mostrar el estado de la máquina
		} else if linea == "--FIN DE RESPUESTA--" {
			fmt.Println("Comando exitoso. Respuesta completa recibida.")
			break
		} else {
			fmt.Println("Servidor dice:", linea)
		}
	}
}

func leerEntrada() string {
	lector := bufio.NewReader(os.Stdin)
	entrada, _ := lector.ReadString('\n')
	return strings.TrimSpace(entrada)
}
