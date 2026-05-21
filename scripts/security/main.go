package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func checkHealth(url string) {
	resp, err := http.Get(url + "/health")
	if err != nil {
		fmt.Println("API caida - no se puede conectar", err)
		return
	}

	if resp.StatusCode == 200 {
		fmt.Println("API viva - Status:", resp.StatusCode)
	} else {
		fmt.Println("API respondio con Status:", resp.StatusCode)
	}

}

func fuzzTitles(url string) {
	payloads := []string{
		"",
		"   ",
		"<script>alert('xss')<script>",
		"'OR 1=1--",
		"../../../etc/passwd",
		"A",
	}
	fmt.Println("\n------Fuzzing titulos-----")

	bloqueados := 0
	aceptados := 0

	for _, payload := range payloads {
		body := map[string]string{
			"title":       payload,
			"description": "test",
		}

		jsonBody, _ := json.Marshal(body)
		resp, err := http.Post(
			url+"/api/v1/tasks",
			"application/json",
			bytes.NewBuffer(jsonBody),
		)

		if err != nil {
			fmt.Printf("X Error enviando %q: %v\n", payload, err)
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode == 201 {
			aceptados++
			fmt.Printf("Payload: %-35q → Status: %d\n", payload, resp.StatusCode)
		} else {
			bloqueados++
			fmt.Printf("Payload: %-35q → Status: %d | %s\n",
				payload, resp.StatusCode, string(respBody))
		}

	}
	fmt.Println("\n--- Resumen ---")
	fmt.Printf("Total payloads : %d\n", len(payloads))
	fmt.Printf("Bloqueados  : %d\n", bloqueados)
	fmt.Printf("Aceptados   : %d\n", aceptados)

}

func testRateLimit(url string) {
	fmt.Println("\n----Rate Limit Test------")

	total := 900
	start := time.Now()

	resultados := make(chan int, total)

	for i := 1; i < total; i++ {
		go func() {
			resp, err := http.Get(url + "/health")
			if err != nil {
				resultados <- 0
				return
			}
			resp.Body.Close()
			resultados <- resp.StatusCode
		}()
	}

	errores := 0
	for i := 1; i < total; i++ {
		status := <-resultados
		if status != 200 {
			errores++
		}
	}

	duracion := time.Since(start)

	fmt.Printf("Total peticiones : %d\n", total)
	fmt.Printf("Errores          : %d\n", errores)
	fmt.Printf("Tiempo total     : %v\n", duracion)
	fmt.Printf("Promedio por req : %v\n", duracion/time.Duration(total))

}

func testHeaderInjection(url string) {
	fmt.Println("\n----Header Injection Test------")
	tests := []struct {
		name   string
		header string
		value  string
	}{
		{"X-Forwarded-For bypass", "X-Forwarded-For", "127.0.0.1"},
		{"X-Real-IP bypass", "X-Real-IP", "localhost"},
		{"Host override", "Host", "evil.com"},
		{"Auth bypass", "Authorization", "Bearer invalid_token"},
		{"Admin bypass", "X-Admin", "true"},
		{"Role escalation", "X-Role", "admin"},
		{"CRLF injection", "X-Custom", "value\r\nX-Injected: malicious"},
	}

	bloqueados := 0
	aceptados := 0

	for _, test := range tests {
		req, err := http.NewRequest("GET", url+"/api/v1/tasks", nil)
		if err != nil {
			fmt.Printf("X Error creando request: %v\n", err)
			continue
		}

		req.Header.Set(test.header, test.value)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("X Error enviando %q: %v\n", test.name, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == 200 || resp.StatusCode == 201 {
			aceptados++
			fmt.Printf("⚠️  %-25s → Status: %d (header aceptado)\n", test.name, resp.StatusCode)
		} else {
			bloqueados++
			fmt.Printf("✅ %-25s → Status: %d (bloqueado)\n", test.name, resp.StatusCode)
		}
	}

	fmt.Println("\n--- Resumen ---")
	fmt.Printf("✅ Bloqueados : %d\n", bloqueados)
	fmt.Printf("⚠️  Aceptados  : %d\n", aceptados)
}

func testHTTPMethods(url string) {
	fmt.Println("\n----HTTP Method Fuzzing------")

	endpoints := []string{
		"/health",
		"/api/v1/tasks",
		"/api/v1/tasks/1",
	}

	methods := []string{
		"GET", "POST", "PUT", "DELETE",
		"PATCH", "OPTIONS", "TRACE", "HEAD",
		"CONNECT", "PROPFIND", "MKCOL",
	}

	for _, endpoint := range endpoints {
		fmt.Printf("\n→ Endpoint: %s\n", endpoint)

		for _, method := range methods {
			req, err := http.NewRequest(method, url+endpoint, nil)
			if err != nil {
				fmt.Printf("  X Error creando request %s: %v\n", method, err)
				continue
			}

			client := &http.Client{}
			resp, err := client.Do(req)
			if err != nil {
				fmt.Printf("  X %-10s → Error: %v\n", method, err)
				continue
			}
			resp.Body.Close()

			switch {
			case resp.StatusCode == 405:
				fmt.Printf("  ✅ %-10s → %d (método no permitido)\n", method, resp.StatusCode)
			case resp.StatusCode == 200 || resp.StatusCode == 201:
				fmt.Printf("  ⚠️  %-10s → %d (aceptado)\n", method, resp.StatusCode)
			case resp.StatusCode == 404:
				fmt.Printf("  🔵 %-10s → %d (not found)\n", method, resp.StatusCode)
			default:
				fmt.Printf("  🟡 %-10s → %d\n", method, resp.StatusCode)
			}
		}
	}
}

func testPathTraversal(url string) {
	fmt.Println("\n----Path Traversal Test------")

	paths := []string{
		// IDs inválidos
		"/api/v1/tasks/0",
		"/api/v1/tasks/-1",
		"/api/v1/tasks/99999999",
		"/api/v1/tasks/abc",
		"/api/v1/tasks/ ",

		// Path traversal clásico
		"/api/v1/tasks/../../../etc/passwd",
		"/api/v1/tasks/%2e%2e%2f%2e%2e%2fetc%2fpasswd",
		"/api/v1/tasks/..%2F..%2Fetc%2Fpasswd",

		// Rutas inexistentes
		"/api/v1/tasks/1/admin",
		"/api/v1/tasks/1/../../users",
		"/api/v1/../admin",
		"/api/v1/tasks/null",
		"/api/v1/tasks/undefined",
		"/api/v1/tasks/NaN",

		// Caracteres especiales
		"/api/v1/tasks/1;drop table tasks",
		"/api/v1/tasks/<script>",
		"/api/v1/tasks/1%00",
	}

	for _, path := range paths {
		req, err := http.NewRequest("GET", url+path, nil)
		if err != nil {
			fmt.Printf("  X Error creando request %q: %v\n", path, err)
			continue
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("  X %-45q → Error: %v\n", path, err)
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		switch {
		case resp.StatusCode == 200:
			fmt.Printf("  ⚠️  %-45q → %d | %s\n", path, resp.StatusCode, string(body))
		case resp.StatusCode == 404 || resp.StatusCode == 400:
			fmt.Printf("  ✅ %-45q → %d\n", path, resp.StatusCode)
		default:
			fmt.Printf("  🟡 %-45q → %d | %s\n", path, resp.StatusCode, string(body))
		}
	}
}

func testLargePayload(url string) {
	fmt.Println("\n----Large Payload Test------")

	tests := []struct {
		name      string
		titleSize int
		descSize  int
	}{
		{"Normal", 100, 200},
		{"Grande", 1_000, 2_000},
		{"Muy grande", 10_000, 20_000},
		{"Enorme", 100_000, 200_000},
		{"Gigante", 1_000_000, 2_000_000},
	}

	for _, test := range tests {
		body := map[string]string{
			"title":       strings.Repeat("A", test.titleSize),
			"description": strings.Repeat("B", test.descSize),
		}

		jsonBody, _ := json.Marshal(body)

		start := time.Now()
		resp, err := http.Post(
			url+"/api/v1/tasks",
			"application/json",
			bytes.NewBuffer(jsonBody),
		)
		duracion := time.Since(start)

		if err != nil {
			fmt.Printf("  X %-15s → Error: %v\n", test.name, err)
			continue
		}

		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		payloadSize := float64(len(jsonBody)) / 1024

		switch {
		case resp.StatusCode == 413:
			fmt.Printf("  ✅ %-15s → %d (payload too large) | %.1f KB | %v\n",
				test.name, resp.StatusCode, payloadSize, duracion)
		case resp.StatusCode == 201:
			fmt.Printf("  ⚠️  %-15s → %d (aceptado)          | %.1f KB | %v\n",
				test.name, resp.StatusCode, payloadSize, duracion)
		case resp.StatusCode == 400 || resp.StatusCode == 422:
			fmt.Printf("  🟡 %-15s → %d (validación)        | %.1f KB | %v | %s\n",
				test.name, resp.StatusCode, payloadSize, duracion, string(respBody))
		default:
			fmt.Printf("  🟡 %-15s → %d                     | %.1f KB | %v\n",
				test.name, resp.StatusCode, payloadSize, duracion)
		}
	}
}

func testBruteForce(url string) {
	fmt.Println("\n----Brute Force Test------")

	passwords := []string{
		"123456", "password", "admin", "admin123",
		"12345678", "qwerty", "abc123", "letmein",
		"monkey", "1234567890", "password1", "iloveyou",
	}

	emails := []string{
		"admin@admin.com",
		"admin@test.com",
		"user@test.com",
	}

	encontrados := 0

	for _, email := range emails {
		for _, password := range passwords {
			body := map[string]string{
				"email":    email,
				"password": password,
			}
			jsonBody, _ := json.Marshal(body)
			resp, err := http.Post(
				url+"/api/v1/auth/login",
				"application/json",
				bytes.NewBuffer(jsonBody),
			)
			if err != nil {
				continue
			}
			respBody, _ := io.ReadAll(resp.Body)
			resp.Body.Close()

			if resp.StatusCode == 200 {
				encontrados++
				fmt.Printf("  🚨 CREDENCIALES ENCONTRADAS: %s / %s\n", email, password)
				fmt.Printf("     Respuesta: %s\n", string(respBody))
			} else {
				fmt.Printf("  ✅ %-30s / %-15s → %d\n", email, password, resp.StatusCode)
			}

			time.Sleep(50 * time.Millisecond)
		}
	}

	fmt.Printf("\n  Total encontrados: %d\n", encontrados)
}

func main() {
	fmt.Println("Iniciando scanner...")
	url := "http://localhost:8080"
	fmt.Println("Target:", url)
	checkHealth(url)
	fuzzTitles(url)
	testRateLimit(url)
	testHeaderInjection(url)
	testHTTPMethods(url)
	testPathTraversal(url)
	testLargePayload(url)
	testBruteForce(url)

}
