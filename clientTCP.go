package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	White  = "\033[37m"
)

func main() {
	// Comprobar que se pasen los argumentos correctamente
	if len(os.Args) < 3 {
		fmt.Println("Uso: go run clientTCP.go <IP> <Puerto>")
		return
	}

	// Leer la IP y el puerto de los argumentos
	ip := os.Args[1]
	puerto := os.Args[2]

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
		fmt.Print("\n%s$ %s", Green, Reset) // Muestra el prompt de la terminal
		comando := leerEntrada()
		imprimirComando(comando)

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

func imprimirComando(comando string) {
	fmt.Printf("%s$ %s%s\n", Green, comando, Reset)
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
			imprimirEstadoMaquina(linea) // Usar formato adecuado para el estado
		} else if linea == "--FIN DE RESPUESTA--" {
			fmt.Println("Comando exitoso. Respuesta completa recibida.")
			break
		} else {
			fmt.Printf("%s[SERVER]: %s%s\n", Yellow, linea, Reset) // Resalta la salida del servidor
		}
	}
}

func imprimirEstadoMaquina(estado string) {
	fmt.Printf("%sEstado de la máquina:%s %s\n", Blue, Reset, estado)
}

func leerEntrada() string {
	lector := bufio.NewReader(os.Stdin)
	entrada, _ := lector.ReadString('\n')
	return strings.TrimSpace(entrada)
}
