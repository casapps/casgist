package email

import (
	"context"
	"log"
	"time"

	"github.com/spf13/viper"
)

// Processor handles background email processing
type Processor struct {
	service *Service
	cfg     *viper.Viper
	stop    chan bool
	stopped chan bool
}

// NewProcessor creates a new email processor
func NewProcessor(service *Service, cfg *viper.Viper) *Processor {
	return &Processor{
		service: service,
		cfg:     cfg,
		stop:    make(chan bool, 1),
		stopped: make(chan bool, 1),
	}
}

// Start begins processing emails in the background
func (p *Processor) Start(ctx context.Context) {
	if !p.cfg.GetBool("email.enabled") {
		log.Println("Email processing is disabled")
		return
	}

	log.Println("Starting email processor...")

	// Process interval (default: 30 seconds)
	interval := p.cfg.GetDuration("email.process_interval")
	if interval == 0 {
		interval = 30 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Email processor stopping due to context cancellation")
			p.stopped <- true
			return
		case <-p.stop:
			log.Println("Email processor stopping...")
			p.stopped <- true
			return
		case <-ticker.C:
			p.processEmails(ctx)
		}
	}
}

// Stop stops the email processor
func (p *Processor) Stop() {
	select {
	case p.stop <- true:
	default:
	}
	
	// Wait for processor to stop
	select {
	case <-p.stopped:
	case <-time.After(5 * time.Second):
		log.Println("Email processor stop timeout")
	}
}

// processEmails processes pending emails
func (p *Processor) processEmails(ctx context.Context) {
	if err := p.service.ProcessEmailQueue(ctx); err != nil {
		log.Printf("Error processing email queue: %v", err)
	}
}