package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"text/template"

	"github.com/MKMuhammetKaradag/go-microservice/email-service/routes"
	"github.com/MKMuhammetKaradag/go-microservice/shared/messaging"
	"github.com/joho/godotenv"
)

type EmailData struct {
	ActivationCode string
	UserName       string
}

func main() {
	config := messaging.NewDefaultConfig()
	config.RetryTypes = []string{"active_user", "forgot_password"}
	rabbit, err := messaging.NewRabbitMQ(config, messaging.EmailService)
	if err != nil {
		log.Fatal("RabbitMQ bağlantı hatası:", err)
	}
	defer rabbit.Close()

	// Mesaj dinleyiciyi başlat
	err = rabbit.ConsumeMessages(func(msg messaging.Message) error {
		if msg.Type == "active_user" || msg.Type == "forgot_password" {
			fmt.Println("active_user  or forgot_password  geldi")
			fmt.Println(msg)
			// return nil
			return handleSendEmail(msg)
		}
		return nil
	})
	if err != nil {
		log.Fatal("Mesaj dinleyici başlatılamadı:", err)
	}
	port := 8082
	fmt.Printf("Auth Service running on port %d\n", port)
	r := routes.CreateServer()
	http.ListenAndServe(fmt.Sprintf(":%d", port), r)
}
func renderTemplate(templatePath string, data EmailData) (string, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// E-posta gönderme fonksiyonu
func sendEmail(subject, body, recipient string) error {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Çevresel değişkenler yüklenemedi:", err)
	}

	from := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")

	auth := smtp.PlainAuth("", from, password, smtpHost)

	msg := []byte("Subject: " + subject + "\r\n" +
		"Content-Type: text/html; charset=\"utf-8\"\r\n" +
		"\r\n" +
		body)

	err = smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{recipient}, msg)
	if err != nil {
		return err
	}
	return nil
}
func handleSendEmail(msg messaging.Message) error {
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		return fmt.Errorf("geçersiz mesaj formatı")
	}

	email, emailOk := data["email"].(string)
	activationCode, codeOk := data["activation_code"].(string)
	templateName, templateOk := data["template_name"].(string)
	userName, userNameOk := data["userName"].(string)

	if !emailOk || !codeOk || !templateOk || !userNameOk {
		log.Printf("Eksik email, aktivasyon kodu veya şablon adı: %+v", data)
	}

	var subject string
	switch msg.Type {
	case "active_user":
		subject = "Hesap Aktivasyonu"
	case "forgot_password":
		subject = "Şifre Sıfırlama"
	default:
		log.Printf("Desteklenmeyen komut: %v", msg.Type)
	}

	// Aktivasyon e-postası için dinamik veriler
	emailData := EmailData{
		ActivationCode: activationCode,
		UserName:       userName,
	}

	// Şablonu oluştur
	body, err := renderTemplate("templates/"+templateName, emailData)
	if err != nil {
		log.Printf("Şablon oluşturulamadı: %v", err)
	}

	// E-posta gönder
	err = sendEmail(subject, body, email)
	if err != nil {
		log.Printf("E-posta gönderilemedi: %v", err)

	}

	log.Printf("E-posta başarıyla gönderildi. Alıcı: %s", email)
	return nil
}
