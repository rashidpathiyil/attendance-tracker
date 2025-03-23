package main

import (
	"encoding/base64"
	"log"

	"fyne.io/fyne/v2"
)

// DEPRECATED: Use resourceIconPng from bundled.go instead
// This function provided for backward compatibility only
func resourceAppIcon() fyne.Resource {
	// Basic checkmark icon encoded in base64
	icon := "iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAACXBIWXMAAAsTAAALEwEAmpwYAAABu0lEQVR4nO2ZMU7DMBRAQ9iYGLkNLEyVOAMSJ2Bk5gSwcQmOwA3gDMAGd0CqxMjSiSuAOEB/xJdwFCVNGuK4tfNGS5Gbn/95f+ykCAAAAAAAuG78QBShN1HkVahDBTpXqGMF2lKoDYVaU6BV48lKKTKNuijQrgIdKtCxAp0p0E2FetOg92PGGUfZvzPQ3CQ5a/1sBboVoJnfKcbD+vDzzUCfOG0m4W8FRv2JgE7qYM3kE7jJhSN+PgGc3VGKtU+UE9D1RU7A3KzN53kQkDkBQ58lSW4C+J5wDgJmyQWGPg/o1ZY7ATvmAFyMwLg/GTClkncBpwhwSPIuIFYBwdfO5l7eOMzjmN83gbHvAsx1O5anwCrJgF7K3CwDV0xFQIdY4wTsmQvsDLxq1NaM4p8OQbsSYL7TbMvrrhFpVjNJJ0ug4wtdB3S0CYq/HW4ykL8R2qhlIKVnZ2w7vLTLz3Tp21iArhUkl8KPy95FpnYt0dDW2OIFmEN99UUB+1UBW+MxVe2Nd3m//pjbeoCJ32aDpvbGKvf+p32A/29hSuNVG/QpPgfz5MHMAAAAAAAAAMCVN38B+AIdPFQFAAAAAElFTkSuQmCC"

	data, err := base64.StdEncoding.DecodeString(icon)
	if err != nil {
		log.Println("Error decoding icon:", err)
		return nil
	}

	return fyne.NewStaticResource("icon.png", data)
}
