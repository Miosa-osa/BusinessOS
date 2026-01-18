package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rhl/businessos-backend/internal/config"
	"github.com/rhl/businessos-backend/internal/services"
	voicev1 "github.com/rhl/businessos-backend/proto/voice/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
)

// VoiceServer runs the gRPC voice service
type VoiceServer struct {
	port            int
	pool            *pgxpool.Pool
	grpcServer      *grpc.Server
	voiceController *services.VoiceController
}

// NewVoiceServer creates a new gRPC voice server
func NewVoiceServer(port int, pool *pgxpool.Pool, cfg *config.Config) *VoiceServer {
	// Create service dependencies
	sttService := services.NewWhisperService()
	ttsService := services.NewElevenLabsService()

	// Create embedding and summarizer services for context
	ollamaURL := os.Getenv("OLLAMA_URL")
	if ollamaURL == "" {
		ollamaURL = "http://localhost:11434"
	}
	embeddingService := services.NewEmbeddingService(pool, ollamaURL)
	summarizerService := services.NewSummarizerService(pool, cfg)

	// Create tiered context service
	contextService := services.NewTieredContextService(pool, embeddingService, summarizerService)

	// Create voice controller
	voiceController := services.NewVoiceController(
		pool,
		sttService,
		ttsService,
		contextService,
	)

	return &VoiceServer{
		port:            port,
		pool:            pool,
		voiceController: voiceController,
	}
}

// Start starts the gRPC server
func (vs *VoiceServer) Start(ctx context.Context) error {
	// Listen on TCP port
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", vs.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Create gRPC server with options
	vs.grpcServer = grpc.NewServer(
		grpc.MaxRecvMsgSize(10*1024*1024), // 10MB for audio chunks
		grpc.MaxSendMsgSize(10*1024*1024), // 10MB for audio chunks
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     15 * time.Minute,
			MaxConnectionAge:      30 * time.Minute,
			MaxConnectionAgeGrace: 5 * time.Second,
			Time:                  5 * time.Second,
			Timeout:               1 * time.Second,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	// Register voice service
	voicev1.RegisterVoiceServiceServer(vs.grpcServer, vs.voiceController)

	// Register reflection service (for grpcurl and debugging)
	reflection.Register(vs.grpcServer)

	slog.Info("[VoiceServer] Starting gRPC server",
		"port", vs.port,
		"address", fmt.Sprintf("0.0.0.0:%d", vs.port))

	// Handle graceful shutdown
	go vs.handleShutdown(ctx)

	// Serve gRPC
	if err := vs.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// handleShutdown handles graceful shutdown
func (vs *VoiceServer) handleShutdown(ctx context.Context) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigCh:
		slog.Info("[VoiceServer] Received shutdown signal", "signal", sig.String())
		vs.Shutdown()
	case <-ctx.Done():
		slog.Info("[VoiceServer] Context cancelled, shutting down")
		vs.Shutdown()
	}
}

// Shutdown gracefully shuts down the gRPC server
func (vs *VoiceServer) Shutdown() {
	slog.Info("[VoiceServer] Shutting down gRPC server...")

	// Graceful stop with timeout
	done := make(chan struct{})
	go func() {
		vs.grpcServer.GracefulStop()
		close(done)
	}()

	// Wait for graceful stop or force stop after 10 seconds
	select {
	case <-done:
		slog.Info("[VoiceServer] Graceful shutdown complete")
	case <-time.After(10 * time.Second):
		slog.Warn("[VoiceServer] Graceful shutdown timeout, forcing stop")
		vs.grpcServer.Stop()
	}
}
