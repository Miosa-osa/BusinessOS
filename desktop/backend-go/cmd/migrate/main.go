package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env
	godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close(ctx)

	// Add share_calendar column
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'team_members' AND column_name = 'share_calendar'
			) THEN
				ALTER TABLE team_members ADD COLUMN share_calendar BOOLEAN DEFAULT FALSE;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding share_calendar: %v", err)
	} else {
		fmt.Println("✓ share_calendar column OK")
	}

	// Add calendar_user_id column
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'team_members' AND column_name = 'calendar_user_id'
			) THEN
				ALTER TABLE team_members ADD COLUMN calendar_user_id VARCHAR(255);
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding calendar_user_id: %v", err)
	} else {
		fmt.Println("✓ calendar_user_id column OK")
	}

	// Add client_id column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'client_id'
			) THEN
				ALTER TABLE projects ADD COLUMN client_id UUID REFERENCES clients(id) ON DELETE SET NULL;
				CREATE INDEX IF NOT EXISTS idx_projects_client ON projects(client_id);
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding client_id: %v", err)
	} else {
		fmt.Println("✓ client_id column OK")
	}

	// Add start_date column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'start_date'
			) THEN
				ALTER TABLE projects ADD COLUMN start_date TIMESTAMPTZ;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding start_date: %v", err)
	} else {
		fmt.Println("✓ start_date column OK")
	}

	// Add end_date column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'end_date'
			) THEN
				ALTER TABLE projects ADD COLUMN end_date TIMESTAMPTZ;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding end_date: %v", err)
	} else {
		fmt.Println("✓ end_date column OK")
	}

	// Add budget column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'budget'
			) THEN
				ALTER TABLE projects ADD COLUMN budget DECIMAL(12,2);
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding budget: %v", err)
	} else {
		fmt.Println("✓ budget column OK")
	}

	// Add due_date column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'due_date'
			) THEN
				ALTER TABLE projects ADD COLUMN due_date TIMESTAMPTZ;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding due_date: %v", err)
	} else {
		fmt.Println("✓ due_date column OK")
	}

	// Add completed_at column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'completed_at'
			) THEN
				ALTER TABLE projects ADD COLUMN completed_at TIMESTAMPTZ;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding completed_at: %v", err)
	} else {
		fmt.Println("✓ completed_at column OK")
	}

	// Add visibility column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'visibility'
			) THEN
				ALTER TABLE projects ADD COLUMN visibility VARCHAR(20) DEFAULT 'private';
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding visibility: %v", err)
	} else {
		fmt.Println("✓ visibility column OK")
	}

	// Add tags column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'tags'
			) THEN
				ALTER TABLE projects ADD COLUMN tags TEXT[];
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding tags: %v", err)
	} else {
		fmt.Println("✓ tags column OK")
	}

	// Add archived column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'archived'
			) THEN
				ALTER TABLE projects ADD COLUMN archived BOOLEAN DEFAULT FALSE;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding archived: %v", err)
	} else {
		fmt.Println("✓ archived column OK")
	}

	// Add progress column to projects
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'progress'
			) THEN
				ALTER TABLE projects ADD COLUMN progress INTEGER DEFAULT 0;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding progress: %v", err)
	} else {
		fmt.Println("✓ progress column OK")
	}

	// Add owner_id column to projects (no FK since users table may not exist)
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name = 'projects' AND column_name = 'owner_id'
			) THEN
				ALTER TABLE projects ADD COLUMN owner_id VARCHAR(255);
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding owner_id: %v", err)
	} else {
		fmt.Println("✓ owner_id column OK")
	}

	// ===== CHAIN OF THOUGHT (COT) THINKING SYSTEM =====

	// Create thinkingtype enum
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'thinkingtype') THEN
				CREATE TYPE thinkingtype AS ENUM ('analysis', 'planning', 'reflection', 'tool_use', 'reasoning', 'evaluation');
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning creating thinkingtype enum: %v", err)
	} else {
		fmt.Println("✓ thinkingtype enum OK")
	}

	// Create reasoning_templates table (must be created before thinking_traces due to FK)
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS reasoning_templates (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			description TEXT,
			system_prompt TEXT,
			thinking_instruction TEXT,
			output_format VARCHAR(50) DEFAULT 'streaming',
			show_thinking BOOLEAN DEFAULT true,
			save_thinking BOOLEAN DEFAULT true,
			max_thinking_tokens INT DEFAULT 4096,
			times_used INT DEFAULT 0,
			is_default BOOLEAN DEFAULT false,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_reasoning_templates_user ON reasoning_templates(user_id);
	`)
	if err != nil {
		log.Printf("Warning creating reasoning_templates table: %v", err)
	} else {
		fmt.Println("✓ reasoning_templates table OK")
	}

	// Create thinking_traces table
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS thinking_traces (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id VARCHAR(255) NOT NULL,
			conversation_id UUID REFERENCES conversations(id) ON DELETE CASCADE,
			message_id UUID REFERENCES messages(id) ON DELETE CASCADE,
			thinking_content TEXT NOT NULL,
			thinking_type thinkingtype DEFAULT 'reasoning',
			step_number INT DEFAULT 1,
			started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			completed_at TIMESTAMP WITH TIME ZONE,
			duration_ms INT,
			thinking_tokens INT DEFAULT 0,
			model_used VARCHAR(100),
			reasoning_template_id UUID REFERENCES reasoning_templates(id) ON DELETE SET NULL,
			metadata JSONB DEFAULT '{}',
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_thinking_traces_user ON thinking_traces(user_id);
		CREATE INDEX IF NOT EXISTS idx_thinking_traces_conversation ON thinking_traces(conversation_id);
		CREATE INDEX IF NOT EXISTS idx_thinking_traces_message ON thinking_traces(message_id);
	`)
	if err != nil {
		log.Printf("Warning creating thinking_traces table: %v", err)
	} else {
		fmt.Println("✓ thinking_traces table OK")
	}

	// Add thinking settings columns to user_settings
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'user_settings' AND column_name = 'thinking_enabled'
			) THEN
				ALTER TABLE user_settings ADD COLUMN thinking_enabled BOOLEAN DEFAULT false;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding thinking_enabled: %v", err)
	} else {
		fmt.Println("✓ thinking_enabled column OK")
	}

	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'user_settings' AND column_name = 'thinking_show_in_ui'
			) THEN
				ALTER TABLE user_settings ADD COLUMN thinking_show_in_ui BOOLEAN DEFAULT true;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding thinking_show_in_ui: %v", err)
	} else {
		fmt.Println("✓ thinking_show_in_ui column OK")
	}

	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'user_settings' AND column_name = 'thinking_save_traces'
			) THEN
				ALTER TABLE user_settings ADD COLUMN thinking_save_traces BOOLEAN DEFAULT true;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding thinking_save_traces: %v", err)
	} else {
		fmt.Println("✓ thinking_save_traces column OK")
	}

	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'user_settings' AND column_name = 'thinking_default_template_id'
			) THEN
				ALTER TABLE user_settings ADD COLUMN thinking_default_template_id UUID REFERENCES reasoning_templates(id) ON DELETE SET NULL;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding thinking_default_template_id: %v", err)
	} else {
		fmt.Println("✓ thinking_default_template_id column OK")
	}

	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'user_settings' AND column_name = 'thinking_max_tokens'
			) THEN
				ALTER TABLE user_settings ADD COLUMN thinking_max_tokens INT DEFAULT 4096;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding thinking_max_tokens: %v", err)
	} else {
		fmt.Println("✓ thinking_max_tokens column OK")
	}

	// ===== CUSTOM AGENTS SYSTEM (Migration 009) =====

	// Create custom_agents table
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS custom_agents (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id VARCHAR(255) NOT NULL,
			name VARCHAR(50) NOT NULL,
			display_name VARCHAR(100) NOT NULL,
			description TEXT,
			avatar VARCHAR(50),
			system_prompt TEXT NOT NULL,
			model_preference VARCHAR(100),
			temperature DECIMAL(3,2) DEFAULT 0.7,
			max_tokens INTEGER DEFAULT 4096,
			capabilities TEXT[] DEFAULT '{}',
			tools_enabled TEXT[] DEFAULT '{}',
			context_sources TEXT[] DEFAULT '{}',
			thinking_enabled BOOLEAN DEFAULT FALSE,
			streaming_enabled BOOLEAN DEFAULT TRUE,
			category VARCHAR(50) DEFAULT 'general',
			is_public BOOLEAN DEFAULT FALSE,
			is_active BOOLEAN DEFAULT TRUE,
			times_used INTEGER DEFAULT 0,
			last_used_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(user_id, name)
		);
		CREATE INDEX IF NOT EXISTS idx_custom_agents_user_id ON custom_agents(user_id);
		CREATE INDEX IF NOT EXISTS idx_custom_agents_name ON custom_agents(user_id, name);
		CREATE INDEX IF NOT EXISTS idx_custom_agents_category ON custom_agents(category);
		CREATE INDEX IF NOT EXISTS idx_custom_agents_active ON custom_agents(user_id, is_active);
	`)
	if err != nil {
		log.Printf("Warning creating custom_agents table: %v", err)
	} else {
		fmt.Println("✓ custom_agents table OK")
	}

	// Create agent_presets table
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS agent_presets (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(50) NOT NULL UNIQUE,
			display_name VARCHAR(100) NOT NULL,
			description TEXT,
			avatar VARCHAR(50),
			system_prompt TEXT NOT NULL,
			model_preference VARCHAR(100),
			temperature DECIMAL(3,2) DEFAULT 0.7,
			max_tokens INTEGER DEFAULT 4096,
			capabilities TEXT[] DEFAULT '{}',
			tools_enabled TEXT[] DEFAULT '{}',
			context_sources TEXT[] DEFAULT '{}',
			thinking_enabled BOOLEAN DEFAULT FALSE,
			category VARCHAR(50) DEFAULT 'general',
			times_copied INTEGER DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`)
	if err != nil {
		log.Printf("Warning creating agent_presets table: %v", err)
	} else {
		fmt.Println("✓ agent_presets table OK")
	}

	// Insert default agent presets
	_, err = conn.Exec(ctx, `
		INSERT INTO agent_presets (name, display_name, description, avatar, system_prompt, category, capabilities, tools_enabled, thinking_enabled)
		VALUES
			('code-reviewer', 'Code Reviewer', 'Reviews code for bugs, best practices, and improvements', 'magnifying-glass',
			 'You are an expert code reviewer. Analyze code for:
1. **Bugs & Errors**: Identify potential bugs, edge cases, and runtime errors
2. **Best Practices**: Check adherence to coding standards and conventions
3. **Performance**: Spot inefficiencies and suggest optimizations
4. **Security**: Flag potential security vulnerabilities
5. **Maintainability**: Assess code readability and suggest improvements

Provide specific, actionable feedback with code examples when suggesting changes.',
			 'coding', ARRAY['code_review', 'analysis'], ARRAY['read_file', 'search_code'], TRUE),

			('technical-writer', 'Technical Writer', 'Creates clear documentation and technical content', 'pencil',
			 'You are an expert technical writer. Create clear, well-structured documentation that:
1. Uses simple, precise language
2. Includes relevant code examples
3. Follows standard documentation patterns
4. Anticipates reader questions
5. Provides both quick-start guides and detailed references

Adapt your writing style to the audience - from beginner-friendly tutorials to expert reference docs.',
			 'writing', ARRAY['documentation', 'writing'], ARRAY[]::TEXT[], FALSE),

			('data-analyst', 'Data Analyst', 'Analyzes data and creates insights', 'chart',
			 'You are an expert data analyst. When analyzing data:
1. Start with exploratory analysis to understand the data
2. Identify key patterns, trends, and anomalies
3. Use appropriate statistical methods
4. Create clear visualizations (describe them in detail)
5. Provide actionable insights and recommendations

Be precise with numbers and transparent about limitations or assumptions.',
			 'analysis', ARRAY['data_analysis', 'visualization'], ARRAY[]::TEXT[], TRUE),

			('business-strategist', 'Business Strategist', 'Provides strategic business advice and analysis', 'briefcase',
			 'You are a senior business strategist. Provide strategic advice by:
1. Understanding the business context and objectives
2. Analyzing market conditions and competition
3. Identifying opportunities and risks
4. Developing actionable recommendations
5. Considering implementation feasibility

Use frameworks like SWOT, Porter''s Five Forces, and business model canvas when appropriate.',
			 'business', ARRAY['strategy', 'analysis', 'planning'], ARRAY[]::TEXT[], TRUE),

			('creative-writer', 'Creative Writer', 'Helps with creative writing and content creation', 'sparkles',
			 'You are a talented creative writer. Help with:
1. Generating creative ideas and concepts
2. Writing engaging narratives and copy
3. Developing compelling characters and stories
4. Crafting persuasive marketing content
5. Editing and improving existing content

Match the desired tone, style, and voice. Be creative while staying on-brand.',
			 'writing', ARRAY['creative_writing', 'content_creation'], ARRAY[]::TEXT[], FALSE)
		ON CONFLICT (name) DO NOTHING;
	`)
	if err != nil {
		log.Printf("Warning inserting agent presets: %v", err)
	} else {
		fmt.Println("✓ agent_presets data OK")
	}

	// ===== CUSTOM COMMANDS SYSTEM (Migration 010) =====

	// Create custom_commands table
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS custom_commands (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id VARCHAR(255) NOT NULL,
			trigger VARCHAR(50) NOT NULL,
			display_name VARCHAR(100) NOT NULL,
			description TEXT,
			action_type VARCHAR(50) NOT NULL,
			target_agent_id UUID REFERENCES custom_agents(id) ON DELETE SET NULL,
			prompt_template TEXT,
			tool_name VARCHAR(100),
			requires_input BOOLEAN DEFAULT FALSE,
			input_placeholder TEXT,
			parameters JSONB DEFAULT '{}',
			streaming_enabled BOOLEAN DEFAULT TRUE,
			thinking_enabled BOOLEAN DEFAULT FALSE,
			category VARCHAR(50) DEFAULT 'general',
			is_active BOOLEAN DEFAULT TRUE,
			is_system BOOLEAN DEFAULT FALSE,
			times_used INTEGER DEFAULT 0,
			last_used_at TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			UNIQUE(user_id, trigger)
		);
		CREATE INDEX IF NOT EXISTS idx_custom_commands_user_id ON custom_commands(user_id);
		CREATE INDEX IF NOT EXISTS idx_custom_commands_trigger ON custom_commands(user_id, trigger);
		CREATE INDEX IF NOT EXISTS idx_custom_commands_active ON custom_commands(user_id, is_active);
		CREATE INDEX IF NOT EXISTS idx_custom_commands_category ON custom_commands(category);
	`)
	if err != nil {
		log.Printf("Warning creating custom_commands table: %v", err)
	} else {
		fmt.Println("✓ custom_commands table OK")
	}

	// Create agent_mentions table
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS agent_mentions (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id VARCHAR(255) NOT NULL,
			conversation_id UUID NOT NULL,
			message_id UUID NOT NULL,
			mentioned_agent_id UUID REFERENCES custom_agents(id) ON DELETE CASCADE,
			mention_text VARCHAR(100) NOT NULL,
			position_in_message INT,
			resolved BOOLEAN DEFAULT TRUE,
			resolution_note TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_agent_mentions_user_id ON agent_mentions(user_id);
		CREATE INDEX IF NOT EXISTS idx_agent_mentions_conversation ON agent_mentions(conversation_id);
		CREATE INDEX IF NOT EXISTS idx_agent_mentions_agent ON agent_mentions(mentioned_agent_id);
	`)
	if err != nil {
		log.Printf("Warning creating agent_mentions table: %v", err)
	} else {
		fmt.Println("✓ agent_mentions table OK")
	}

	// Insert default system commands
	_, err = conn.Exec(ctx, `
		INSERT INTO custom_commands (user_id, trigger, display_name, description, action_type, prompt_template, category, is_system, streaming_enabled, thinking_enabled)
		VALUES
			('SYSTEM', '/help', 'Show Help', 'Display available commands and agents', 'template',
			 'Here are the available commands and agents:\n\n**Commands:**\n{{commands_list}}\n\n**Agents:**\n{{agents_list}}',
			 'productivity', TRUE, TRUE, FALSE),

			('SYSTEM', '/clear', 'Clear Context', 'Clear conversation context', 'tool',
			 NULL,
			 'productivity', TRUE, FALSE, FALSE),

			('SYSTEM', '/summarize', 'Summarize Conversation', 'Create a summary of the current conversation', 'template',
			 'Please provide a concise summary of this conversation, highlighting:\n1. Key topics discussed\n2. Decisions made\n3. Action items identified\n4. Open questions remaining',
			 'productivity', TRUE, TRUE, TRUE)
		ON CONFLICT (user_id, trigger) DO NOTHING;
	`)
	if err != nil {
		log.Printf("Warning inserting system commands: %v", err)
	} else {
		fmt.Println("✓ system_commands data OK")
	}

	// ===== CORE SPECIALIST AGENTS (Migration 011) =====

	// Insert 5 core specialist agent presets
	_, err = conn.Exec(ctx, `
		INSERT INTO agent_presets (name, display_name, description, avatar, system_prompt, model_preference, category, capabilities, tools_enabled, thinking_enabled, temperature)
		VALUES
			-- RESEARCHER
			('researcher', 'Researcher', 'Expert at deep research, fact-finding, and synthesizing information from multiple sources', 'search',
			 'You are an expert Research Agent specializing in deep investigation and knowledge synthesis.

## Core Capabilities
- **Deep Research**: Thoroughly investigate topics using all available sources
- **Fact Verification**: Cross-reference information to ensure accuracy
- **Source Attribution**: Always cite sources and provide references
- **Knowledge Synthesis**: Combine information from multiple sources into coherent insights

## Research Process
1. **Understand the Query**: Clarify what information is needed
2. **Gather Sources**: Search for relevant, authoritative sources
3. **Analyze & Verify**: Cross-check facts across multiple sources
4. **Synthesize**: Combine findings into clear, actionable insights
5. **Cite**: Always provide sources for claims

## Output Format
- Start with a brief summary of findings
- Provide detailed analysis with citations
- Include confidence levels for claims
- Suggest areas for further research if needed

Always prioritize accuracy over speed. If uncertain, say so explicitly.',
			 'claude-3-5-sonnet-20241022', 'research',
			 ARRAY['research', 'fact_checking', 'synthesis', 'web_search'],
			 ARRAY['web_search', 'read_document'],
			 TRUE, 0.3),

			-- WRITER
			('writer', 'Writer', 'Professional writer for all types of content - articles, emails, documentation, and creative writing', 'pen-tool',
			 'You are an expert Writing Agent capable of creating high-quality content across all formats.

## Core Capabilities
- **Versatile Writing**: Articles, emails, docs, marketing copy, creative content
- **Tone Adaptation**: Match any voice, style, or brand guidelines
- **Structure**: Clear organization with compelling flow
- **Editing**: Polish and improve existing content

## Writing Process
1. **Understand the Brief**: Clarify purpose, audience, tone, and constraints
2. **Outline**: Create a logical structure before writing
3. **Draft**: Write with the target audience in mind
4. **Refine**: Edit for clarity, flow, and impact
5. **Polish**: Final review for grammar and style

## Output Guidelines
- Match the requested tone and style
- Use clear, engaging language
- Structure content for easy reading (headers, bullets, short paragraphs)
- Provide multiple versions if requested

Adapt your writing style to the context - professional for business, engaging for marketing, precise for technical.',
			 NULL, 'writing',
			 ARRAY['content_creation', 'copywriting', 'editing', 'documentation'],
			 ARRAY[]::TEXT[],
			 FALSE, 0.7),

			-- CODER
			('coder', 'Coder', 'Expert software developer for writing, reviewing, and debugging code across all languages', 'code',
			 'You are an expert Coding Agent with deep knowledge of software development.

## Core Capabilities
- **Multi-Language**: Proficient in all major programming languages
- **Code Writing**: Write clean, efficient, well-documented code
- **Code Review**: Identify bugs, security issues, and improvements
- **Debugging**: Systematic problem diagnosis and fixes
- **Architecture**: Design scalable, maintainable solutions

## Coding Standards
1. **Clean Code**: Readable, self-documenting, follows conventions
2. **Error Handling**: Robust error handling and edge cases
3. **Security**: No vulnerabilities (OWASP top 10 awareness)
4. **Performance**: Efficient algorithms and data structures
5. **Testing**: Include tests when appropriate

## Output Format
- Provide complete, working code
- Include comments for complex logic
- Explain your approach when asked
- Suggest alternatives when relevant

Always prioritize code quality and security. Ask for clarification on requirements when needed.',
			 'claude-3-5-sonnet-20241022', 'coding',
			 ARRAY['code_writing', 'code_review', 'debugging', 'architecture'],
			 ARRAY['read_file', 'write_file', 'execute_code', 'search_code'],
			 TRUE, 0.2),

			-- ANALYST
			('analyst', 'Analyst', 'Data analyst for extracting insights, identifying patterns, and making data-driven recommendations', 'bar-chart-2',
			 'You are an expert Analysis Agent specializing in data interpretation and insights.

## Core Capabilities
- **Data Analysis**: Statistical analysis, trend identification, pattern recognition
- **Visualization**: Describe effective charts and visualizations
- **Interpretation**: Transform raw data into actionable insights
- **Forecasting**: Predictive analysis and projections
- **Reporting**: Clear, executive-ready summaries

## Analysis Framework
1. **Understand the Question**: What decision needs to be made?
2. **Explore the Data**: Initial patterns, distributions, quality
3. **Analyze**: Apply appropriate statistical methods
4. **Interpret**: What do the numbers mean?
5. **Recommend**: Actionable next steps

## Output Guidelines
- Lead with key findings and recommendations
- Support claims with specific data points
- Visualize (describe charts) when helpful
- Acknowledge data limitations and assumptions
- Provide confidence intervals when appropriate

Be precise with numbers. Distinguish between correlation and causation.',
			 NULL, 'analysis',
			 ARRAY['data_analysis', 'statistics', 'visualization', 'forecasting'],
			 ARRAY['read_file', 'execute_code'],
			 TRUE, 0.3),

			-- PLANNER
			('planner', 'Planner', 'Strategic planner for project planning, goal setting, and creating actionable roadmaps', 'map',
			 'You are an expert Planning Agent specializing in strategy and project management.

## Core Capabilities
- **Strategic Planning**: Long-term vision and goal setting
- **Project Planning**: Breakdown, timelines, dependencies
- **Resource Allocation**: Optimize people, time, and budget
- **Risk Management**: Identify and mitigate potential issues
- **Progress Tracking**: Milestones and success metrics

## Planning Framework
1. **Define Goals**: Clear, measurable objectives (SMART)
2. **Analyze Context**: Current state, constraints, resources
3. **Develop Strategy**: High-level approach to achieve goals
4. **Create Roadmap**: Phases, milestones, dependencies
5. **Plan Execution**: Detailed tasks, owners, deadlines

## Output Format
- Start with executive summary
- Provide clear timelines and milestones
- Include task breakdowns with dependencies
- Identify risks and mitigation strategies
- Define success metrics and checkpoints

Focus on actionable, realistic plans. Consider constraints and dependencies.',
			 NULL, 'planning',
			 ARRAY['strategic_planning', 'project_management', 'roadmapping', 'risk_assessment'],
			 ARRAY[]::TEXT[],
			 TRUE, 0.5)
		ON CONFLICT (name) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			description = EXCLUDED.description,
			avatar = EXCLUDED.avatar,
			system_prompt = EXCLUDED.system_prompt,
			model_preference = EXCLUDED.model_preference,
			category = EXCLUDED.category,
			capabilities = EXCLUDED.capabilities,
			tools_enabled = EXCLUDED.tools_enabled,
			thinking_enabled = EXCLUDED.thinking_enabled,
			temperature = EXCLUDED.temperature,
			updated_at = NOW();
	`)
	if err != nil {
		log.Printf("Warning inserting core specialist presets: %v", err)
	} else {
		fmt.Println("✓ core_specialist_presets data OK")
	}

	// Migration 012: Add thinking_tokens to usage tracking
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			-- Add thinking_tokens column to ai_usage_logs
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'ai_usage_logs' AND column_name = 'thinking_tokens'
			) THEN
				ALTER TABLE ai_usage_logs ADD COLUMN thinking_tokens INTEGER DEFAULT 0;
			END IF;

			-- Add ai_thinking_tokens to usage_daily_summary
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'usage_daily_summary' AND column_name = 'ai_thinking_tokens'
			) THEN
				ALTER TABLE usage_daily_summary ADD COLUMN ai_thinking_tokens BIGINT DEFAULT 0;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding thinking_tokens columns: %v", err)
	} else {
		fmt.Println("✓ thinking_tokens columns OK")
	}

	// Migration 013: Focus Configurations System
	_, err = conn.Exec(ctx, `
		-- Focus mode templates (system-level defaults)
		CREATE TABLE IF NOT EXISTS focus_mode_templates (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			name VARCHAR(50) NOT NULL UNIQUE,
			display_name VARCHAR(100) NOT NULL,
			description TEXT,
			icon VARCHAR(50),
			default_model VARCHAR(100),
			temperature DECIMAL(3,2) DEFAULT 0.7,
			max_tokens INTEGER DEFAULT 4096,
			output_style VARCHAR(50) DEFAULT 'balanced',
			response_format VARCHAR(50) DEFAULT 'markdown',
			max_response_length INTEGER,
			require_sources BOOLEAN DEFAULT false,
			auto_search BOOLEAN DEFAULT false,
			search_depth VARCHAR(20) DEFAULT 'quick',
			kb_context_limit INTEGER DEFAULT 5,
			include_history_count INTEGER DEFAULT 10,
			thinking_enabled BOOLEAN DEFAULT false,
			thinking_style VARCHAR(50),
			system_prompt_prefix TEXT,
			system_prompt_suffix TEXT,
			sort_order INTEGER DEFAULT 0,
			is_active BOOLEAN DEFAULT true,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);

		-- User-specific focus configuration overrides
		CREATE TABLE IF NOT EXISTS focus_configurations (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id VARCHAR(255) NOT NULL,
			template_id UUID REFERENCES focus_mode_templates(id) ON DELETE CASCADE,
			custom_name VARCHAR(100),
			temperature DECIMAL(3,2),
			max_tokens INTEGER,
			output_style VARCHAR(50),
			response_format VARCHAR(50),
			max_response_length INTEGER,
			require_sources BOOLEAN,
			auto_search BOOLEAN,
			search_depth VARCHAR(20),
			kb_context_limit INTEGER,
			include_history_count INTEGER,
			thinking_enabled BOOLEAN,
			thinking_style VARCHAR(50),
			custom_system_prompt TEXT,
			preferred_model VARCHAR(100),
			auto_load_kb_categories TEXT[],
			keyboard_shortcut VARCHAR(20),
			is_favorite BOOLEAN DEFAULT false,
			use_count INTEGER DEFAULT 0,
			last_used_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW(),
			UNIQUE(user_id, template_id)
		);

		-- Focus context presets
		CREATE TABLE IF NOT EXISTS focus_context_presets (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			user_id VARCHAR(255) NOT NULL,
			name VARCHAR(100) NOT NULL,
			description TEXT,
			kb_artifact_ids UUID[],
			kb_categories TEXT[],
			project_ids UUID[],
			default_search_queries TEXT[],
			search_domains TEXT[],
			created_at TIMESTAMPTZ DEFAULT NOW(),
			updated_at TIMESTAMPTZ DEFAULT NOW()
		);

		-- Link presets to focus configurations
		CREATE TABLE IF NOT EXISTS focus_configuration_presets (
			focus_config_id UUID REFERENCES focus_configurations(id) ON DELETE CASCADE,
			preset_id UUID REFERENCES focus_context_presets(id) ON DELETE CASCADE,
			sort_order INTEGER DEFAULT 0,
			PRIMARY KEY (focus_config_id, preset_id)
		);

		-- Indexes
		CREATE INDEX IF NOT EXISTS idx_focus_configurations_user ON focus_configurations(user_id);
		CREATE INDEX IF NOT EXISTS idx_focus_configurations_template ON focus_configurations(template_id);
		CREATE INDEX IF NOT EXISTS idx_focus_context_presets_user ON focus_context_presets(user_id);
	`)
	if err != nil {
		log.Printf("Warning creating focus_mode tables: %v", err)
	} else {
		fmt.Println("✓ focus_mode tables OK")
	}

	// Insert default focus mode templates
	_, err = conn.Exec(ctx, `
		INSERT INTO focus_mode_templates (name, display_name, description, icon, default_model, temperature, output_style, auto_search, search_depth, thinking_enabled, thinking_style, system_prompt_prefix, sort_order) VALUES
		('quick', 'Quick', 'Fast, concise responses for simple questions', 'zap', NULL, 0.5, 'concise', false, 'quick', false, NULL, 'You are in Quick Mode. Provide brief, direct answers. Be concise and to the point.', 1),
		('deep', 'Deep Research', 'Thorough research with sources and citations', 'search', 'claude-sonnet-4-20250514', 0.7, 'detailed', true, 'deep', true, 'analytical', 'You are in Deep Research Mode. Conduct thorough research and provide comprehensive, well-sourced answers.', 2),
		('creative', 'Creative', 'Imaginative and exploratory responses', 'sparkles', NULL, 0.9, 'balanced', false, 'quick', true, 'creative', 'You are in Creative Mode. Think outside the box. Be imaginative and innovative.', 3),
		('analyze', 'Analysis', 'Data-driven analysis and insights', 'chart-bar', 'claude-sonnet-4-20250514', 0.6, 'structured', false, 'standard', true, 'analytical', 'You are in Analysis Mode. Focus on data-driven insights with clear structure.', 4),
		('write', 'Writing', 'Document creation and editing', 'file-text', NULL, 0.7, 'detailed', false, 'quick', false, NULL, 'You are in Writing Mode. Create well-structured, polished content.', 5),
		('plan', 'Planning', 'Strategic planning and project organization', 'clipboard-list', NULL, 0.6, 'structured', false, 'standard', true, 'step-by-step', 'You are in Planning Mode. Create actionable plans with clear steps.', 6),
		('code', 'Coding', 'Software development assistance', 'code', 'claude-sonnet-4-20250514', 0.4, 'structured', false, 'quick', true, 'step-by-step', 'You are in Coding Mode. Write clean, efficient code with best practices.', 7)
		ON CONFLICT (name) DO UPDATE SET
			display_name = EXCLUDED.display_name,
			description = EXCLUDED.description,
			updated_at = NOW();
	`)
	if err != nil {
		log.Printf("Warning inserting focus_mode_templates: %v", err)
	} else {
		fmt.Println("✓ focus_mode_templates data OK")
	}

	// ===== Migration 014: Web Search Results Cache =====
	_, err = conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS web_search_results (
			id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
			query_hash VARCHAR(64) NOT NULL,
			original_query TEXT NOT NULL,
			optimized_query TEXT,
			user_id VARCHAR(255),
			conversation_id UUID,
			results JSONB NOT NULL DEFAULT '[]',
			result_count INTEGER DEFAULT 0,
			provider VARCHAR(50) DEFAULT 'duckduckgo',
			search_time_ms FLOAT,
			expires_at TIMESTAMPTZ NOT NULL,
			hit_count INTEGER DEFAULT 0,
			last_hit_at TIMESTAMPTZ,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_web_search_query_hash ON web_search_results(query_hash);
		CREATE INDEX IF NOT EXISTS idx_web_search_user ON web_search_results(user_id) WHERE user_id IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_web_search_conversation ON web_search_results(conversation_id) WHERE conversation_id IS NOT NULL;
		CREATE INDEX IF NOT EXISTS idx_web_search_expires ON web_search_results(expires_at);
		CREATE INDEX IF NOT EXISTS idx_web_search_lookup ON web_search_results(query_hash, expires_at);
	`)
	if err != nil {
		log.Printf("Warning creating web_search_results table: %v", err)
	} else {
		fmt.Println("✓ web_search_results table OK")
	}

	// Add is_active column to agent_presets if missing
	_, err = conn.Exec(ctx, `
		DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_name = 'agent_presets' AND column_name = 'is_active'
			) THEN
				ALTER TABLE agent_presets ADD COLUMN is_active BOOLEAN DEFAULT TRUE;
			END IF;
		END $$;
	`)
	if err != nil {
		log.Printf("Warning adding is_active to agent_presets: %v", err)
	} else {
		fmt.Println("✓ agent_presets.is_active column OK")
	}

	fmt.Println("Migration complete!")
}
