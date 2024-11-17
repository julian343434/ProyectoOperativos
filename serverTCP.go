package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"gopkg.in/ini.v1"
)

// Mutex para sincronización
var mutex sync.Mutex

// Variables de configuración
var puerto int
var ipClientePermitida string
var intentosMaximosAutenticacion int

func main() {
	cargarConfiguracion() // Cargar configuración desde archivo
	imprimirBanner()
	iniciarServidor()
}

// Cargar configuración desde el archivo server.config
func cargarConfiguracion() {
	cfg, err := ini.Load("server.config")
	if err != nil {
		fmt.Println("Error cargando archivo de configuración:", err)
		os.Exit(1)
	}

	// Leer la configuración
	ipClientePermitida = cfg.Section("").Key("ip_cliente_permitida").String()
	puerto, _ = cfg.Section("").Key("puerto").Int()
	intentosMaximosAutenticacion, _ = cfg.Section("").Key("intentos_maximos_autenticacion").Int()
}

func imprimirBanner() {
	fmt.Println("**")
	fmt.Println("*       Server TCP Operativos          *")
	fmt.Println("**")
	fmt.Printf("Abriendo el Puerto %d...\n", puerto)
}

func iniciarServidor() {
	tcpAddress, err := net.ResolveTCPAddr("tcp4", fmt.Sprintf("%s:%d", ipClientePermitida, puerto))
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
	fmt.Printf("Servidor escuchando en %s:%d...\n", ipClientePermitida, puerto)

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

func iniciarEstadoPeriodico(socket net.Conn, duracion time.Duration, detenerEstado chan bool) {
	ticker := time.NewTicker(duracion) // Crea el ticker para el envío periódico
	defer ticker.Stop()

	for {
		select {
		case <-detenerEstado: // Si recibe la señal de detención, finaliza la goroutine
			fmt.Println("Deteniendo envío de estado de la máquina...")
			return
		case <-ticker.C: // Enviar el estado de la máquina en intervalos definidos
			reporte := obtenerConsumoRecursos()
			socket.Write([]byte("Estado de la máquina: " + reporte + "\n"))
		}
	}
}
func manejarConexion(socket net.Conn) {
	defer socket.Close() // Cierra la conexión cuando termine
	reader := bufio.NewReader(socket)

	// Enviar el número de intentos máximos al cliente
	socket.Write([]byte(fmt.Sprintf("INTENTOS:%d\n", intentosMaximosAutenticacion)))

	// Autenticación
	if !autenticarUsuario(reader, socket) {
		socket.Write([]byte("Error de autenticación.\n"))
		return
	}
	socket.Write([]byte("OK\n"))

	// Leer el tiempo del cliente para la duración del estado
	tiempo, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error leyendo tiempo:", err)
		return
	}
	tiempo = strings.TrimSpace(strings.TrimPrefix(tiempo, "TIEMPO:"))

	// Convertir el tiempo en duración
	duracion, err := parsearDuracion(tiempo)
	if err != nil {
		fmt.Println("Formato de tiempo inválido:", tiempo)
		socket.Write([]byte("Formato de tiempo inválido.\n"))
		return
	}

	// Canal para detener el estado periódico
	detenerEstado := make(chan bool)

	// Goroutine para enviar estado de la máquina en intervalos definidos
	go func() {
		iniciarEstadoPeriodico(socket, duracion, detenerEstado)
	}()

	// Bucle principal para manejar comandos
	for {
		socket.Write([]byte("Servidor listo para recibir comandos.\n"))
		comando, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error leyendo comando:", err)
			break
		}
		comando = strings.TrimSpace(comando)

		// Manejo del comando 'exit'
		if comando == "exit" {
			fmt.Println("Cliente desconectado:", socket.RemoteAddr())
			socket.Write([]byte("--FIN DE RESPUESTA--\n"))
			detenerEstado <- true // Detener el envío periódico
			break
		}

		// Ejecutar comando y enviar la respuesta
		salida, err := ejecutarComando(comando)
		if err != nil {
			socket.Write([]byte(fmt.Sprintf("Error ejecutando comando: %s\n", err)))
		} else {
			enviarRespuesta(socket, salida)
		}
		socket.Write([]byte("--FIN DE RESPUESTA--\n"))
	}

	// Esperar a que la goroutine de estado termine antes de salir de la función
	<-detenerEstado // Esperar a que se reciba la señal de detención del estado
	fmt.Println("Conexión cerrada correctamente.")
}

func autenticarUsuario(reader *bufio.Reader, socket net.Conn) bool {
	for intentos := 0; intentos < intentosMaximosAutenticacion; intentos++ {
		authData, _ := reader.ReadString('\n')
		authParts := strings.Split(strings.TrimSpace(authData), ":")

		if len(authParts) != 2 {
			socket.Write([]byte("Error en formato de autenticación.\n"))
			continue
		}

		usuario, contrasena := authParts[0], authParts[1]
		if verificarCredenciales(usuario, contrasena) {
			return true
		} else {
			socket.Write([]byte(fmt.Sprintf("Autenticación fallida (%d/%d intentos).\n", intentos+1, intentosMaximosAutenticacion)))
		}
	}
	return false
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

func parsearDuracion(tiempo string) (time.Duration, error) {
	if len(tiempo) < 2 {
		return 0, fmt.Errorf("tiempo inválido")
	}

	unidad := tiempo[len(tiempo)-1]
	valor := tiempo[:len(tiempo)-1]
	duracion, err := strconv.Atoi(valor)
	if err != nil {
		return 0, err
	}

	switch unidad {
	case 's': // Segundos
		return time.Duration(duracion) * time.Second, nil
	case 'm': // Minutos
		return time.Duration(duracion) * time.Minute, nil
	case 'h': // Horas
		return time.Duration(duracion) * time.Hour, nil
	default:
		return 0, fmt.Errorf("unidad no reconocida")
	}
}
