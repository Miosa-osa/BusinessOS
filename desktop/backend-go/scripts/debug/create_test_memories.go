package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	ctx := context.Background()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	workspaceID := uuid.MustParse("064e8e2a-5d3e-4d00-8492-df3628b1ec96")
	userID := "ZVtQRaictVbO9lN0p-csSA"

	fmt.Println("Creating test memories...")

	// Memory 1: About Pedro (workspace visible, pinned)
	mem1ID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_memories (
			id, workspace_id, owner_user_id, title, summary, content,
			visibility, memory_type, importance_score, is_pinned,
			created_by, created_at, updated_at
		) VALUES (
			$1, $2, NULL, $3, $4, $5,
			$6, $7, $8, $9,
			$10, NOW(), NOW()
		)
	`, mem1ID, workspaceID, "Informação sobre Pedro",
		"Pedro é o desenvolvedor backend do BusinessOS",
		"Pedro é o desenvolvedor backend do BusinessOS. Ele trabalha principalmente com Go e PostgreSQL. É especialista em otimização de banco de dados e arquitetura de microsserviços.",
		"workspace", "fact", 0.9, true, userID)

	if err != nil {
		log.Printf("Failed to create memory 1: %v", err)
	} else {
		fmt.Println("✓ Created: Informação sobre Pedro (workspace, pinned)")
	}

	// Memory 2: About BusinessOS (workspace visible)
	mem2ID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_memories (
			id, workspace_id, owner_user_id, title, summary, content,
			visibility, memory_type, importance_score, is_pinned,
			created_by, created_at, updated_at
		) VALUES (
			$1, $2, NULL, $3, $4, $5,
			$6, $7, $8, $9,
			$10, NOW(), NOW()
		)
	`, mem2ID, workspaceID, "Sobre o BusinessOS",
		"BusinessOS é uma plataforma de gerenciamento empresarial completa",
		"BusinessOS é uma plataforma de gerenciamento empresarial completa. Inclui gestão de projetos, clientes, equipe, calendário e um sistema de chat com IA. Foi desenvolvido usando SvelteKit no frontend e Go no backend.",
		"workspace", "project_info", 0.8, false, userID)

	if err != nil {
		log.Printf("Failed to create memory 2: %v", err)
	} else {
		fmt.Println("✓ Created: Sobre o BusinessOS (workspace)")
	}

	// Memory 3: Private note for user (private)
	mem3ID := uuid.New()
	_, err = pool.Exec(ctx, `
		INSERT INTO workspace_memories (
			id, workspace_id, owner_user_id, title, summary, content,
			visibility, memory_type, importance_score, is_pinned,
			created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6,
			$7, $8, $9, $10,
			$11, NOW(), NOW()
		)
	`, mem3ID, workspaceID, userID, "Nota Privada de Teste",
		"Nota privada de teste",
		"Esta é uma memória privada visível apenas para mim.",
		"private", "note", 0.5, false, userID)

	if err != nil {
		log.Printf("Failed to create memory 3: %v", err)
	} else {
		fmt.Println("✓ Created: Nota Privada de Teste (private)")
	}

	fmt.Println("\nDone! Test memories created.")
	fmt.Println("\nNow refresh the chat and send a message like 'Quem é Pedro?' or 'O que é o BusinessOS?'")
	fmt.Println("The agent should include information from these memories in the response!")
}
