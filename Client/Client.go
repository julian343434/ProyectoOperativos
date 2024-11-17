package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	// Verificar que los argumentos sean suficientes
	if len(os.Args) < 4 {
		fmt.Println("Faltan argumentos. Uso: clienteOperativos <IP> <Puerto> <PeriodoReporte>")
		return
	}

	// Obtener los argumentos
	ip := os.Args[1]
	puerto := os.Args[2]
	periodoReporte := os.Args[3]

	// Intentar convertir el periodo de reporte a tiempo de duración
	duration, err := time.ParseDuration(periodoReporte)
	if err != nil {
		fmt.Println("Error al convertir el periodo de reporte:", err)
		return
	}

	// Mostrar los valores obtenidos
	fmt.Println("Conectando al servidor en IP:", ip, "y puerto:", puerto)
	fmt.Println("Periodo de reporte:", duration)

	// Puedes agregar el código que use la duración aquí
}
