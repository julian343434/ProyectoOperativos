package main

import (
	"bufio"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// Config estructura para almacenar la configuración
type Config struct {
	IPPermitida      string
	Puerto           string
	IntentosFallidos int
}

// Leer configuración desde el archivo .conf
func leerConfiguracion(ruta string) Config {
	file, err := os.Open(ruta)
	if err != nil {
		fmt.Println("Error leyendo archivo de configuración:", err)
		os.Exit(1)
	}
	defer file.Close()

	config := Config{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		linea := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(linea, "ip_permitida") {
			config.IPPermitida = strings.Split(linea, "=")[1]
		} else if strings.HasPrefix(linea, "puerto") {
			config.Puerto = strings.Split(linea, "=")[1]
		} else if strings.HasPrefix(linea, "intentos_fallidos") {
			fmt.Sscanf(strings.Split(linea, "=")[1], "%d", &config.IntentosFallidos)
		}
	}
	return config
}

// Hash de contraseña usando SHA-256
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// Validar usuario en la base de datos
func validarUsuario(username, password string) bool {
	db, err := sql.Open("sqlite3", "./usuarios.db")
	if err != nil {
		fmt.Println("Error abriendo base de datos:", err)
		return false
	}
	defer db.Close()

	var hashedPassword string
	err = db.QueryRow("SELECT password FROM usuarios WHERE username = ?", username).Scan(&hashedPassword)
	if err != nil {
		fmt.Println("Usuario no encontrado:", username)
		return false
	}

	return hashedPassword == password
}

// Iniciar el servidor
func iniciarServidor(config Config) {
	listener, err := net.Listen("tcp", ":"+config.Puerto)
	if err != nil {
		fmt.Println("Error iniciando servidor:", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Servidor escuchando en %s:%s...\n", config.IPPermitida, config.Puerto)
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error aceptando conexión:", err)
			continue
		}
		go manejarConexion(conn, config)
	}
}

// Manejar conexión con un cliente
func manejarConexion(conn net.Conn, config Config) {
	defer conn.Close()

	reader := bufio.NewReader(conn)
	ipRemota := conn.RemoteAddr().String()
	if !strings.HasPrefix(ipRemota, config.IPPermitida) {
		fmt.Println("Conexión rechazada de IP no permitida:", ipRemota)
		conn.Write([]byte("IP no autorizada\n"))
		return
	}

	if !autenticarCliente(conn, reader) {
		fmt.Println("Autenticación fallida")
		return
	}

	for {
		comando, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error leyendo comando:", err)
			break
		}
		comando = strings.TrimSpace(comando)

		if comando == "bye" {
			fmt.Println("Cliente desconectado")
			break
		}

		salida, err := ejecutarComando(comando)
		if err != nil {
			conn.Write([]byte(fmt.Sprintf("Error ejecutando comando: %s\n", err)))
			continue
		}
		conn.Write([]byte(salida + "\n--FIN DE RESPUESTA--\n"))
	}
}

// Autenticar cliente
func autenticarCliente(conn net.Conn, reader *bufio.Reader) bool {
	conn.Write([]byte("Ingrese sus credenciales (usuario:password):\n"))
	credenciales, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error leyendo credenciales:", err)
		return false
	}

	credenciales = strings.TrimSpace(credenciales)
	parts := strings.Split(credenciales, ":")
	if len(parts) != 2 {
		conn.Write([]byte("Formato de credenciales inválido\n"))
		return false
	}

	username, password := parts[0], hashPassword(parts[1])
	return validarUsuario(username, password)
}

// Ejecutar comando en el sistema
func ejecutarComando(comando string) (string, error) {
	cmd := exec.Command("sh", "-c", comando)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func main() {
	// Leer configuración del archivo
	config := leerConfiguracion("./Server.conf")

	// Iniciar el servidor
	iniciarServidor(config)
}
