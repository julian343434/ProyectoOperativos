package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

func main() {
	imprimirBanner()
	iniciarServidor()
}

func imprimirBanner() {
	fmt.Println("**")
	fmt.Println("*       Server TCP Operativos          *")
	fmt.Println("**")
	fmt.Println("Abriendo el Puerto 2024...")
}

func iniciarServidor() {
	tcpAddress, err := net.ResolveTCPAddr("tcp4", ":2024")
	if err != nil {
		fmt.Println("Error resolviendo dirección:", err)
		return
	}

	serverSocket, err := net.ListenTCP("tcp", tcpAddress)
	if err != nil {
		fmt.Println("Error iniciando servidor:", err)
		return
	}
	defer serverSocket.Close()
	fmt.Println("Servidor escuchando en el puerto 2024...")

	for {
		socket, err := serverSocket.Accept()
		if err != nil {
			fmt.Println("Error aceptando conexión:", err)
			continue
		}
		fmt.Println("Cliente conectado!", socket.RemoteAddr())
		go manejarConexion(socket)
	}
}

func manejarConexion(socket net.Conn) {
	defer socket.Close()
	reader := bufio.NewReader(socket)

	// Autenticación
	if !autenticarUsuario(reader, socket) {
		socket.Write([]byte("Error de autenticación.\n"))
		return
	}
	socket.Write([]byte("OK\n"))

	for {
		socket.Write([]byte("Servidor listo para recibir comandos.\n"))
		comando, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error leyendo comando:", err)
			break
		}
		comando = strings.TrimSpace(comando)

		if comando == "exit" {
			fmt.Println("Cliente desconectado:", socket.RemoteAddr())
			socket.Write([]byte("--FIN DE RESPUESTA--\n"))
			break
		}

		fmt.Println("Ejecutando comando:", comando)
		salida, err := ejecutarComando(comando)
		if err != nil {
			socket.Write([]byte(fmt.Sprintf("Error ejecutando comando: %s\n", err)))
			socket.Write([]byte("--FIN DE RESPUESTA--\n"))
			continue
		}

		// Enviar respuesta del comando
		enviarRespuesta(socket, salida)

		// Enviar el estado de la máquina al cliente
		reporte := obtenerConsumoRecursos()
		socket.Write([]byte("Estado de la máquina: " + reporte + "\n"))
		socket.Write([]byte("--FIN DE RESPUESTA--\n"))
	}
}

func autenticarUsuario(reader *bufio.Reader, socket net.Conn) bool {
	authData, _ := reader.ReadString('\n')
	authParts := strings.Split(strings.TrimSpace(authData), ":")

	if len(authParts) != 2 {
		return false
	}

	usuario, contrasena := authParts[0], authParts[1]
	return verificarCredenciales(usuario, contrasena)
}

func verificarCredenciales(usuario, contrasena string) bool {
	file, _ := os.Open("usuarios.txt")
	defer file.Close()
	scanner := bufio.NewScanner(file)

	hashedPassword := hashSha256(contrasena)

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ":")
		if len(parts) == 2 && parts[0] == usuario && parts[1] == hashedPassword {
			return true
		}
	}
	return false
}

func hashSha256(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func ejecutarComando(comando string) (string, error) {
	cmd := exec.Command("sh", "-c", comando)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func enviarRespuesta(socket net.Conn, salida string) {
	writer := bufio.NewWriter(socket)
	chunks := strings.Split(salida, "\n")
	for _, linea := range chunks {
		writer.WriteString(linea + "\n")
		writer.Flush()
	}
}

func obtenerConsumoRecursos() string {
	usoCPU, _ := cpu.Percent(0, false)
	usoMemoria, _ := mem.VirtualMemory()
	usoDisco, _ := disk.Usage("/")

	return fmt.Sprintf(
		"CPU: %.2f%%, Memoria: %.2f%%, Disco: %.2f%%",
		usoCPU[0], usoMemoria.UsedPercent, usoDisco.UsedPercent,
	)
}