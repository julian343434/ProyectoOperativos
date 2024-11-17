package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
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
	// Verificar argumentos
	if len(os.Args) < 4 {
		fmt.Println("Uso: go run clientTCP.go <IP> <Puerto> <Intervalo>")
		return
	}

	ip := os.Args[1]
	puerto := os.Args[2]
	intervalo := os.Args[3] // Nuevo parámetro

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

	// Enviar intervalo al servidor
	conn.Write([]byte(intervalo + "\n"))

	// Bucle principal para enviar comandos
	for {
		fmt.Printf("\n%s$ %s", Green, Reset) // Muestra el prompt de la terminal
		comando := leerEntrada()

		conn.Write([]byte(comando + "\n"))

		if strings.TrimSpace(comando) == "exit" {
			fmt.Println("Cerrando conexión...")
			break
		}

		recibirRespuesta(conn)

	}
}

func autenticar(conn net.Conn) bool {
	// Leer el número de intentos permitidos desde el servidor
	reader := bufio.NewReader(conn)
	attemptsStr, _ := reader.ReadString('\n')
	attemptsStr = strings.TrimSpace(strings.TrimPrefix(attemptsStr, "INTENTOS:"))
	attempts, err := strconv.Atoi(attemptsStr)
	if err != nil {
		fmt.Println("Error leyendo el número de intentos:", err)
		return false
	}

	// Intentar autenticarse con el número de intentos proporcionado
	for intentos := 0; intentos < attempts; intentos++ {
		fmt.Print("Usuario: ")
		usuario := leerEntrada()
		fmt.Print("Contraseña: ")
		contrasena := leerEntrada()

		authData := usuario + ":" + contrasena
		conn.Write([]byte(authData + "\n"))

		// Leer respuesta de autenticación
		respuesta, _ := reader.ReadString('\n')

		if strings.TrimSpace(respuesta) == "OK" {
			fmt.Println("Autenticación exitosa.")
			return true
		}

		fmt.Printf("Autenticación fallida. Intento %d/%d.\n", intentos+1, attempts)
	}

	return false
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
