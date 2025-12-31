<script lang="ts">
	import { onMount, onDestroy } from 'svelte';
	import { browser } from '$app/environment';

	export type AnimationEffect =
		// Basic
		'none' | 'particles' | 'gradient' | 'pulse' | 'ripples' | 'dots' | 'floatingShapes' | 'smoke' |
		// Nature
		'aurora' | 'fireflies' | 'rain' | 'snow' | 'nebula' | 'waves' | 'bubbles' |
		// Tech
		'starfield' | 'matrix' | 'circuit' | 'confetti' | 'geometric' | 'scanlines' | 'grid' | 'warp' | 'hexgrid' | 'binary';
	export type AnimationIntensity = 'subtle' | 'medium' | 'high';

	interface Props {
		effectType?: AnimationEffect;
		intensity?: AnimationIntensity;
		colors?: string[];
		speed?: number;
	}

	let {
		effectType = 'none',
		intensity = 'subtle',
		colors = ['#667eea', '#764ba2'],
		speed = 1
	}: Props = $props();

	let canvas: HTMLCanvasElement;
	let ctx: CanvasRenderingContext2D | null = null;
	let animationId: number;

	// Particle types
	interface Particle {
		x: number;
		y: number;
		size: number;
		speedX: number;
		speedY: number;
		opacity: number;
		color: string;
	}

	interface Star {
		x: number;
		y: number;
		size: number;
		brightness: number;
		twinkleSpeed: number;
		twinklePhase: number;
	}

	interface MatrixDrop {
		x: number;
		y: number;
		speed: number;
		chars: string[];
		length: number;
		opacity: number;
	}

	interface Bubble {
		x: number;
		y: number;
		size: number;
		speed: number;
		wobble: number;
		wobbleSpeed: number;
		opacity: number;
		color: string;
	}

	interface GeoShape {
		x: number;
		y: number;
		size: number;
		rotation: number;
		rotationSpeed: number;
		speedX: number;
		speedY: number;
		sides: number;
		color: string;
		opacity: number;
	}

	interface Firefly {
		x: number;
		y: number;
		size: number;
		speedX: number;
		speedY: number;
		glowPhase: number;
		glowSpeed: number;
		color: string;
	}

	interface Raindrop {
		x: number;
		y: number;
		length: number;
		speed: number;
		opacity: number;
	}

	interface Snowflake {
		x: number;
		y: number;
		size: number;
		speed: number;
		wobble: number;
		wobbleSpeed: number;
		opacity: number;
	}

	interface ConfettiPiece {
		x: number;
		y: number;
		size: number;
		speedY: number;
		speedX: number;
		rotation: number;
		rotationSpeed: number;
		color: string;
		opacity: number;
	}

	interface Ripple {
		x: number;
		y: number;
		radius: number;
		maxRadius: number;
		speed: number;
		opacity: number;
		color: string;
	}

	interface CircuitNode {
		x: number;
		y: number;
		connections: number[];
		pulsePhase: number;
		pulseSpeed: number;
	}

	interface Dot {
		x: number;
		y: number;
		baseSize: number;
		phase: number;
		speed: number;
	}

	interface FloatingShape {
		x: number;
		y: number;
		size: number;
		rotation: number;
		rotSpeed: number;
		speedX: number;
		speedY: number;
		type: 'square' | 'triangle' | 'circle';
		color: string;
		opacity: number;
	}

	interface SmokeParticle {
		x: number;
		y: number;
		size: number;
		opacity: number;
		speedX: number;
		speedY: number;
		life: number;
	}

	interface GridLine {
		pos: number;
		speed: number;
		opacity: number;
	}

	interface WarpStar {
		x: number;
		y: number;
		z: number;
		prevX: number;
		prevY: number;
	}

	interface HexCell {
		x: number;
		y: number;
		phase: number;
		speed: number;
	}

	let particles: Particle[] = [];
	let stars: Star[] = [];
	let matrixDrops: MatrixDrop[] = [];
	let bubbles: Bubble[] = [];
	let geoShapes: GeoShape[] = [];
	let fireflies: Firefly[] = [];
	let raindrops: Raindrop[] = [];
	let snowflakes: Snowflake[] = [];
	let confetti: ConfettiPiece[] = [];
	let ripples: Ripple[] = [];
	let circuitNodes: CircuitNode[] = [];
	let dots: Dot[] = [];
	let floatingShapes: FloatingShape[] = [];
	let smokeParticles: SmokeParticle[] = [];
	let gridLinesH: GridLine[] = [];
	let gridLinesV: GridLine[] = [];
	let warpStars: WarpStar[] = [];
	let hexCells: HexCell[] = [];
	let scanlineOffset = 0;
	let pulsePhase = 0;
	let nebulaTime = 0;

	// Matrix characters
	const matrixChars = 'アイウエオカキクケコサシスセソタチツテトナニヌネノハヒフヘホマミムメモヤユヨラリルレロワヲン0123456789ABCDEF'.split('');

	// Intensity multipliers
	const intensityConfig = {
		subtle: { particleCount: 30, opacity: 0.3, speed: 0.5 },
		medium: { particleCount: 60, opacity: 0.5, speed: 1 },
		high: { particleCount: 100, opacity: 0.7, speed: 1.5 }
	};

	function initCanvas() {
		if (!browser || !canvas) return;
		ctx = canvas.getContext('2d');
		resizeCanvas();
	}

	function resizeCanvas() {
		if (!canvas) return;
		canvas.width = window.innerWidth;
		canvas.height = window.innerHeight;
		initEffect();
	}

	function initEffect() {
		const config = intensityConfig[intensity];

		switch (effectType) {
			case 'particles':
				initParticles(config.particleCount);
				break;
			case 'starfield':
				initStars(config.particleCount * 2);
				break;
			case 'matrix':
				initMatrix(config.particleCount);
				break;
			case 'bubbles':
				initBubbles(config.particleCount);
				break;
			case 'geometric':
				initGeometric(Math.floor(config.particleCount / 3));
				break;
			case 'fireflies':
				initFireflies(Math.floor(config.particleCount / 2));
				break;
			case 'rain':
				initRain(config.particleCount * 3);
				break;
			case 'snow':
				initSnow(config.particleCount * 2);
				break;
			case 'confetti':
				initConfetti(config.particleCount);
				break;
			case 'ripples':
				initRipples(Math.floor(config.particleCount / 5));
				break;
			case 'circuit':
				initCircuit(Math.floor(config.particleCount / 2));
				break;
			case 'dots':
				initDots(config.particleCount);
				break;
			case 'floatingShapes':
				initFloatingShapes(Math.floor(config.particleCount / 2));
				break;
			case 'smoke':
				initSmoke(Math.floor(config.particleCount / 2));
				break;
			case 'scanlines':
				// Scanlines don't need particle init
				break;
			case 'grid':
				initGrid(Math.floor(config.particleCount / 4));
				break;
			case 'warp':
				initWarp(config.particleCount * 2);
				break;
			case 'hexgrid':
				initHexgrid(Math.floor(config.particleCount / 3));
				break;
			case 'binary':
				initMatrix(Math.floor(config.particleCount / 2)); // Reuse matrix with different render
				break;
		}
	}

	function initParticles(count: number) {
		particles = [];
		for (let i = 0; i < count; i++) {
			particles.push({
				x: Math.random() * canvas.width,
				y: Math.random() * canvas.height,
				size: Math.random() * 3 + 1,
				speedX: (Math.random() - 0.5) * 0.5 * speed,
				speedY: (Math.random() - 0.5) * 0.5 * speed,
				opacity: Math.random() * 0.5 + 0.2,
				color: colors[Math.floor(Math.random() * colors.length)]
			});
		}
	}

	function initStars(count: number) {
		stars = [];
		for (let i = 0; i < count; i++) {
			stars.push({
				x: Math.random() * canvas.width,
				y: Math.random() * canvas.height,
				size: Math.random() * 2 + 0.5,
				brightness: Math.random(),
				twinkleSpeed: Math.random() * 2 + 1,
				twinklePhase: Math.random() * Math.PI * 2
			});
		}
	}

	function initMatrix(count: number) {
		matrixDrops = [];
		const columns = Math.floor(canvas.width / 20);
		const dropsPerColumn = Math.max(1, Math.floor(count / columns));

		for (let col = 0; col < columns; col++) {
			for (let d = 0; d < dropsPerColumn; d++) {
				const length = Math.floor(Math.random() * 15) + 5;
				const chars: string[] = [];
				for (let i = 0; i < length; i++) {
					chars.push(matrixChars[Math.floor(Math.random() * matrixChars.length)]);
				}
				matrixDrops.push({
					x: col * 20 + 10,
					y: Math.random() * canvas.height - canvas.height,
					speed: Math.random() * 2 + 1,
					chars,
					length,
					opacity: Math.random() * 0.5 + 0.3
				});
			}
		}
	}

	function initBubbles(count: number) {
		bubbles = [];
		for (let i = 0; i < count; i++) {
			bubbles.push({
				x: Math.random() * canvas.width,
				y: canvas.height + Math.random() * 100,
				size: Math.random() * 20 + 10,
				speed: Math.random() * 1.5 + 0.5,
				wobble: Math.random() * Math.PI * 2,
				wobbleSpeed: Math.random() * 0.02 + 0.01,
				opacity: Math.random() * 0.3 + 0.1,
				color: colors[Math.floor(Math.random() * colors.length)]
			});
		}
	}

	function initGeometric(count: number) {
		geoShapes = [];
		for (let i = 0; i < count; i++) {
			geoShapes.push({
				x: Math.random() * canvas.width,
				y: Math.random() * canvas.height,
				size: Math.random() * 40 + 20,
				rotation: Math.random() * Math.PI * 2,
				rotationSpeed: (Math.random() - 0.5) * 0.02,
				speedX: (Math.random() - 0.5) * 0.5,
				speedY: (Math.random() - 0.5) * 0.5,
				sides: Math.floor(Math.random() * 4) + 3, // 3-6 sides
				color: colors[Math.floor(Math.random() * colors.length)],
				opacity: Math.random() * 0.3 + 0.1
			});
		}
	}

	function initFireflies(count: number) {
		fireflies = [];
		for (let i = 0; i < count; i++) {
			fireflies.push({
				x: Math.random() * canvas.width,
				y: Math.random() * canvas.height,
				size: Math.random() * 4 + 2,
				speedX: (Math.random() - 0.5) * 0.5,
				speedY: (Math.random() - 0.5) * 0.5,
				glowPhase: Math.random() * Math.PI * 2,
				glowSpeed: Math.random() * 0.03 + 0.01,
				color: colors[Math.floor(Math.random() * colors.length)] || '#ffff00'
			});
		}
	}

	function initRain(count: number) {
		raindrops = [];
		for (let i = 0; i < count; i++) {
			raindrops.push({
				x: Math.random() * canvas.width,
				y: Math.random() * canvas.height,
				length: Math.random() * 20 + 10,
				speed: Math.random() * 10 + 5,
				opacity: Math.random() * 0.3 + 0.2
			});
		}
	}

	function initSnow(count: number) {
		snowflakes = [];
		for (let i = 0; i < count; i++) {
			snowflakes.push({
				x: Math.random() * canvas.width,
				y: Math.random() * canvas.height,
				size: Math.random() * 4 + 1,
				speed: Math.random() * 1 + 0.5,
				wobble: Math.random() * Math.PI * 2,
				wobbleSpeed: Math.random() * 0.02 + 0.01,
				opacity: Math.random() * 0.6 + 0.4
			});
		}
	}

	function initConfetti(count: number) {
		confetti = [];
		const confettiColors = ['#ff6b6b', '#4ecdc4', '#ffe66d', '#a855f7', '#3b82f6', '#22c55e'];
		for (let i = 0; i < count; i++) {
			confetti.push({
				x: Math.random() * canvas.width,
				y: Math.random() * canvas.height - canvas.height,
				size: Math.random() * 8 + 4,
				speedY: Math.random() * 2 + 1,
				speedX: (Math.random() - 0.5) * 2,
				rotation: Math.random() * Math.PI * 2,
				rotationSpeed: (Math.random() - 0.5) * 0.2,
				color: confettiColors[Math.floor(Math.random() * confettiColors.length)],
				opacity: Math.random() * 0.5 + 0.5
			});
		}
	}

	function initRipples(count: number) {
		ripples = [];
		for (let i = 0; i < count; i++) {
			ripples.push({
				x: Math.random() * canvas.width,
				y: Math.random() * canvas.height,
				radius: Math.random() * 20,
				maxRadius: Math.random() * 100 + 50,
				speed: Math.random() * 1 + 0.5,
				opacity: Math.random() * 0.3 + 0.2,
				color: colors[Math.floor(Math.random() * colors.length)]
			});
		}
	}

	function initCircuit(count: number) {
		circuitNodes = [];
		// Ensure minimum node count for visible effect
		const nodeCount = Math.max(count, 20);
		const gridSize = Math.ceil(Math.sqrt(nodeCount));
		const cellWidth = canvas.width / gridSize;
		const cellHeight = canvas.height / gridSize;

		for (let i = 0; i < nodeCount; i++) {
			const gridX = i % gridSize;
			const gridY = Math.floor(i / gridSize);
			circuitNodes.push({
				x: gridX * cellWidth + cellWidth / 2 + (Math.random() - 0.5) * cellWidth * 0.3,
				y: gridY * cellHeight + cellHeight / 2 + (Math.random() - 0.5) * cellHeight * 0.3,
				connections: [],
				pulsePhase: Math.random() * Math.PI * 2,
				pulseSpeed: Math.random() * 0.03 + 0.01
			});
		}

		// Calculate max connection distance based on grid - connect to neighbors
		const maxDist = Math.max(cellWidth, cellHeight) * 1.8;

		// Create connections to nearby nodes
		circuitNodes.forEach((node, index) => {
			const maxConnections = Math.floor(Math.random() * 3) + 2;

			// Sort other nodes by distance and connect to closest ones
			const distances = circuitNodes
				.map((other, otherIndex) => ({
					index: otherIndex,
					dist: Math.sqrt((other.x - node.x) ** 2 + (other.y - node.y) ** 2)
				}))
				.filter(d => d.index !== index && d.dist < maxDist)
				.sort((a, b) => a.dist - b.dist);

			for (let c = 0; c < Math.min(maxConnections, distances.length); c++) {
				if (!node.connections.includes(distances[c].index)) {
					node.connections.push(distances[c].index);
				}
			}
		});
	}

	function initDots(count: number) {
		dots = [];
		const cols = Math.ceil(Math.sqrt(count * 1.5));
		const rows = Math.ceil(count / cols);
		const spacingX = canvas.width / cols;
		const spacingY = canvas.height / rows;

		for (let row = 0; row < rows; row++) {
			for (let col = 0; col < cols; col++) {
				dots.push({
					x: col * spacingX + spacingX / 2,
					y: row * spacingY + spacingY / 2,
					baseSize: 3,
					phase: Math.random() * Math.PI * 2,
					speed: Math.random() * 0.02 + 0.01
				});
			}
		}
	}

	function initFloatingShapes(count: number) {
		floatingShapes = [];
		const shapeTypes: ('square' | 'triangle' | 'circle')[] = ['square', 'triangle', 'circle'];
		for (let i = 0; i < count; i++) {
			floatingShapes.push({
				x: Math.random() * canvas.width,
				y: Math.random() * canvas.height,
				size: Math.random() * 30 + 15,
				rotation: Math.random() * Math.PI * 2,
				rotSpeed: (Math.random() - 0.5) * 0.02,
				speedX: (Math.random() - 0.5) * 0.5,
				speedY: (Math.random() - 0.5) * 0.5,
				type: shapeTypes[Math.floor(Math.random() * shapeTypes.length)],
				color: colors[Math.floor(Math.random() * colors.length)],
				opacity: Math.random() * 0.3 + 0.1
			});
		}
	}

	function initSmoke(count: number) {
		smokeParticles = [];
		for (let i = 0; i < count; i++) {
			smokeParticles.push({
				x: Math.random() * canvas.width,
				y: canvas.height + Math.random() * 50,
				size: Math.random() * 50 + 30,
				opacity: Math.random() * 0.2 + 0.1,
				speedX: (Math.random() - 0.5) * 0.5,
				speedY: -(Math.random() * 0.5 + 0.3),
				life: 1
			});
		}
	}

	function initGrid(count: number) {
		gridLinesH = [];
		gridLinesV = [];
		const spacing = Math.max(canvas.width, canvas.height) / count;

		for (let i = 0; i < Math.ceil(canvas.height / spacing) + 1; i++) {
			gridLinesH.push({
				pos: i * spacing,
				speed: Math.random() * 0.5 + 0.2,
				opacity: Math.random() * 0.3 + 0.1
			});
		}
		for (let i = 0; i < Math.ceil(canvas.width / spacing) + 1; i++) {
			gridLinesV.push({
				pos: i * spacing,
				speed: Math.random() * 0.5 + 0.2,
				opacity: Math.random() * 0.3 + 0.1
			});
		}
	}

	function initWarp(count: number) {
		warpStars = [];
		for (let i = 0; i < count; i++) {
			warpStars.push({
				x: Math.random() * canvas.width - canvas.width / 2,
				y: Math.random() * canvas.height - canvas.height / 2,
				z: Math.random() * 1000,
				prevX: 0,
				prevY: 0
			});
		}
	}

	function initHexgrid(count: number) {
		hexCells = [];
		const hexSize = 30;
		const hexWidth = hexSize * 2;
		const hexHeight = Math.sqrt(3) * hexSize;

		for (let row = 0; row < canvas.height / hexHeight + 1; row++) {
			for (let col = 0; col < canvas.width / (hexWidth * 0.75) + 1; col++) {
				hexCells.push({
					x: col * hexWidth * 0.75,
					y: row * hexHeight + (col % 2) * hexHeight / 2,
					phase: Math.random() * Math.PI * 2,
					speed: Math.random() * 0.02 + 0.01
				});
			}
		}
	}

	function drawParticles() {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		particles.forEach(particle => {
			ctx!.beginPath();
			ctx!.arc(particle.x, particle.y, particle.size, 0, Math.PI * 2);
			ctx!.fillStyle = particle.color + Math.floor(particle.opacity * config.opacity * 255).toString(16).padStart(2, '0');
			ctx!.fill();

			particle.x += particle.speedX * speed;
			particle.y += particle.speedY * speed;

			if (particle.x < 0) particle.x = canvas.width;
			if (particle.x > canvas.width) particle.x = 0;
			if (particle.y < 0) particle.y = canvas.height;
			if (particle.y > canvas.height) particle.y = 0;
		});
	}

	function drawGradient(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		const angle = (time * 0.0001 * speed) % (Math.PI * 2);
		const x1 = canvas.width / 2 + Math.cos(angle) * canvas.width / 2;
		const y1 = canvas.height / 2 + Math.sin(angle) * canvas.height / 2;
		const x2 = canvas.width / 2 + Math.cos(angle + Math.PI) * canvas.width / 2;
		const y2 = canvas.height / 2 + Math.sin(angle + Math.PI) * canvas.height / 2;

		const gradient = ctx.createLinearGradient(x1, y1, x2, y2);
		colors.forEach((color, i) => {
			gradient.addColorStop(i / (colors.length - 1), color);
		});

		ctx.globalAlpha = config.opacity * 0.3;
		ctx.fillStyle = gradient;
		ctx.fillRect(0, 0, canvas.width, canvas.height);
		ctx.globalAlpha = 1;
	}

	function drawAurora(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		for (let layer = 0; layer < 3; layer++) {
			ctx.beginPath();
			ctx.moveTo(0, canvas.height);

			const waveHeight = canvas.height * 0.4;
			const baseY = canvas.height * 0.3 + layer * 50;

			for (let x = 0; x <= canvas.width; x += 5) {
				const wave1 = Math.sin((x * 0.01 + time * 0.0005 * speed) + layer) * waveHeight * 0.3;
				const wave2 = Math.sin((x * 0.02 + time * 0.0003 * speed) + layer * 2) * waveHeight * 0.2;
				const y = baseY + wave1 + wave2;
				ctx.lineTo(x, y);
			}

			ctx.lineTo(canvas.width, canvas.height);
			ctx.closePath();

			const gradient = ctx.createLinearGradient(0, baseY - waveHeight, 0, canvas.height);
			const color = colors[layer % colors.length];
			gradient.addColorStop(0, color + '00');
			gradient.addColorStop(0.5, color + Math.floor(config.opacity * 100).toString(16).padStart(2, '0'));
			gradient.addColorStop(1, color + '00');

			ctx.fillStyle = gradient;
			ctx.fill();
		}
	}

	function drawStarfield(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		stars.forEach(star => {
			const twinkle = Math.sin(time * 0.001 * star.twinkleSpeed * speed + star.twinklePhase);
			const brightness = (star.brightness * 0.5 + twinkle * 0.5 + 0.5) * config.opacity;

			ctx!.beginPath();
			ctx!.arc(star.x, star.y, star.size, 0, Math.PI * 2);
			ctx!.fillStyle = `rgba(255, 255, 255, ${brightness})`;
			ctx!.fill();

			if (star.size > 1.5 && brightness > 0.5) {
				ctx!.beginPath();
				ctx!.arc(star.x, star.y, star.size * 2, 0, Math.PI * 2);
				ctx!.fillStyle = `rgba(255, 255, 255, ${brightness * 0.2})`;
				ctx!.fill();
			}
		});
	}

	function drawMatrix(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		// Slight fade effect for trails
		ctx.fillStyle = 'rgba(0, 0, 0, 0.05)';
		ctx.fillRect(0, 0, canvas.width, canvas.height);

		const fontSize = 14;
		ctx.font = `${fontSize}px 'Courier New', monospace`;

		matrixDrops.forEach(drop => {
			// Draw each character in the drop
			for (let i = 0; i < drop.chars.length; i++) {
				const y = drop.y - i * fontSize;
				if (y < 0 || y > canvas.height + fontSize) continue;

				// Fade based on position in drop
				const fadeRatio = 1 - (i / drop.chars.length);
				const alpha = fadeRatio * drop.opacity * config.opacity;

				// First char is brighter (white/green)
				if (i === 0) {
					ctx!.fillStyle = `rgba(180, 255, 180, ${alpha})`;
				} else {
					const green = Math.floor(100 + fadeRatio * 155);
					ctx!.fillStyle = `rgba(0, ${green}, 70, ${alpha})`;
				}

				// Randomly change characters
				if (Math.random() < 0.02) {
					drop.chars[i] = matrixChars[Math.floor(Math.random() * matrixChars.length)];
				}

				ctx!.fillText(drop.chars[i], drop.x, y);
			}

			// Move drop down
			drop.y += drop.speed * speed * 3;

			// Reset if off screen
			if (drop.y - drop.length * fontSize > canvas.height) {
				drop.y = -drop.length * fontSize;
				drop.x = Math.floor(Math.random() * (canvas.width / 20)) * 20 + 10;
			}
		});
	}

	function drawWaves(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		// Draw multiple wave layers
		for (let layer = 0; layer < 4; layer++) {
			ctx.beginPath();

			const amplitude = 30 + layer * 15;
			const frequency = 0.01 - layer * 0.002;
			const yOffset = canvas.height * 0.5 + layer * 40;
			const phaseOffset = layer * 0.5;

			ctx.moveTo(0, canvas.height);

			for (let x = 0; x <= canvas.width; x += 2) {
				const y = yOffset +
					Math.sin(x * frequency + time * 0.001 * speed + phaseOffset) * amplitude +
					Math.sin(x * frequency * 2 + time * 0.0015 * speed + phaseOffset) * amplitude * 0.5;
				ctx.lineTo(x, y);
			}

			ctx.lineTo(canvas.width, canvas.height);
			ctx.closePath();

			const gradient = ctx.createLinearGradient(0, yOffset - amplitude, 0, canvas.height);
			const color = colors[layer % colors.length] || '#3b82f6';
			const opacityHex = Math.floor(config.opacity * 60 / (layer + 1)).toString(16).padStart(2, '0');
			gradient.addColorStop(0, color + opacityHex);
			gradient.addColorStop(1, color + '00');

			ctx.fillStyle = gradient;
			ctx.fill();
		}
	}

	function drawBubbles(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		bubbles.forEach(bubble => {
			// Wobble effect
			bubble.wobble += bubble.wobbleSpeed * speed;
			const wobbleX = Math.sin(bubble.wobble) * 20;

			// Draw bubble
			ctx!.beginPath();
			ctx!.arc(bubble.x + wobbleX, bubble.y, bubble.size, 0, Math.PI * 2);

			// Gradient fill for 3D effect
			const gradient = ctx!.createRadialGradient(
				bubble.x + wobbleX - bubble.size * 0.3,
				bubble.y - bubble.size * 0.3,
				bubble.size * 0.1,
				bubble.x + wobbleX,
				bubble.y,
				bubble.size
			);
			gradient.addColorStop(0, `rgba(255, 255, 255, ${bubble.opacity * config.opacity * 0.5})`);
			gradient.addColorStop(0.5, bubble.color + Math.floor(bubble.opacity * config.opacity * 150).toString(16).padStart(2, '0'));
			gradient.addColorStop(1, bubble.color + '00');

			ctx!.fillStyle = gradient;
			ctx!.fill();

			// Highlight
			ctx!.beginPath();
			ctx!.arc(
				bubble.x + wobbleX - bubble.size * 0.3,
				bubble.y - bubble.size * 0.3,
				bubble.size * 0.15,
				0, Math.PI * 2
			);
			ctx!.fillStyle = `rgba(255, 255, 255, ${bubble.opacity * config.opacity * 0.6})`;
			ctx!.fill();

			// Move bubble up
			bubble.y -= bubble.speed * speed;

			// Reset if off screen
			if (bubble.y + bubble.size < 0) {
				bubble.y = canvas.height + bubble.size;
				bubble.x = Math.random() * canvas.width;
			}
		});
	}

	function drawGeometric(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		geoShapes.forEach(shape => {
			ctx!.save();
			ctx!.translate(shape.x, shape.y);
			ctx!.rotate(shape.rotation);

			// Draw polygon
			ctx!.beginPath();
			for (let i = 0; i < shape.sides; i++) {
				const angle = (i / shape.sides) * Math.PI * 2 - Math.PI / 2;
				const x = Math.cos(angle) * shape.size;
				const y = Math.sin(angle) * shape.size;
				if (i === 0) {
					ctx!.moveTo(x, y);
				} else {
					ctx!.lineTo(x, y);
				}
			}
			ctx!.closePath();

			// Stroke with gradient effect
			ctx!.strokeStyle = shape.color + Math.floor(shape.opacity * config.opacity * 255).toString(16).padStart(2, '0');
			ctx!.lineWidth = 2;
			ctx!.stroke();

			// Light fill
			ctx!.fillStyle = shape.color + Math.floor(shape.opacity * config.opacity * 50).toString(16).padStart(2, '0');
			ctx!.fill();

			ctx!.restore();

			// Update rotation and position
			shape.rotation += shape.rotationSpeed * speed;
			shape.x += shape.speedX * speed;
			shape.y += shape.speedY * speed;

			// Wrap around edges
			if (shape.x < -shape.size) shape.x = canvas.width + shape.size;
			if (shape.x > canvas.width + shape.size) shape.x = -shape.size;
			if (shape.y < -shape.size) shape.y = canvas.height + shape.size;
			if (shape.y > canvas.height + shape.size) shape.y = -shape.size;
		});
	}

	function drawFireflies(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		fireflies.forEach(fly => {
			fly.glowPhase += fly.glowSpeed * speed;
			const glow = (Math.sin(fly.glowPhase) + 1) / 2;
			const alpha = glow * config.opacity;

			// Draw glow
			const gradient = ctx!.createRadialGradient(fly.x, fly.y, 0, fly.x, fly.y, fly.size * 3);
			gradient.addColorStop(0, fly.color + Math.floor(alpha * 255).toString(16).padStart(2, '0'));
			gradient.addColorStop(1, fly.color + '00');

			ctx!.beginPath();
			ctx!.arc(fly.x, fly.y, fly.size * 3, 0, Math.PI * 2);
			ctx!.fillStyle = gradient;
			ctx!.fill();

			// Draw core
			ctx!.beginPath();
			ctx!.arc(fly.x, fly.y, fly.size, 0, Math.PI * 2);
			ctx!.fillStyle = `rgba(255, 255, 200, ${alpha})`;
			ctx!.fill();

			// Move firefly
			fly.x += fly.speedX * speed;
			fly.y += fly.speedY * speed;

			// Slight random direction change
			if (Math.random() < 0.02) {
				fly.speedX += (Math.random() - 0.5) * 0.2;
				fly.speedY += (Math.random() - 0.5) * 0.2;
				fly.speedX = Math.max(-1, Math.min(1, fly.speedX));
				fly.speedY = Math.max(-1, Math.min(1, fly.speedY));
			}

			// Wrap around
			if (fly.x < 0) fly.x = canvas.width;
			if (fly.x > canvas.width) fly.x = 0;
			if (fly.y < 0) fly.y = canvas.height;
			if (fly.y > canvas.height) fly.y = 0;
		});
	}

	function drawRain(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		raindrops.forEach(drop => {
			ctx!.beginPath();
			ctx!.moveTo(drop.x, drop.y);
			ctx!.lineTo(drop.x + 1, drop.y + drop.length);
			ctx!.strokeStyle = `rgba(150, 180, 220, ${drop.opacity * config.opacity})`;
			ctx!.lineWidth = 1;
			ctx!.stroke();

			// Move drop
			drop.y += drop.speed * speed;
			drop.x += 0.5 * speed; // Slight wind

			// Reset if off screen
			if (drop.y > canvas.height) {
				drop.y = -drop.length;
				drop.x = Math.random() * canvas.width;
			}
		});
	}

	function drawSnow(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		snowflakes.forEach(flake => {
			flake.wobble += flake.wobbleSpeed * speed;
			const wobbleX = Math.sin(flake.wobble) * 20;

			ctx!.beginPath();
			ctx!.arc(flake.x + wobbleX, flake.y, flake.size, 0, Math.PI * 2);
			ctx!.fillStyle = `rgba(255, 255, 255, ${flake.opacity * config.opacity})`;
			ctx!.fill();

			// Move flake
			flake.y += flake.speed * speed;
			flake.x += Math.sin(time * 0.001) * 0.2 * speed;

			// Reset if off screen
			if (flake.y > canvas.height + flake.size) {
				flake.y = -flake.size;
				flake.x = Math.random() * canvas.width;
			}
		});
	}

	function drawConfetti(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		confetti.forEach(piece => {
			ctx!.save();
			ctx!.translate(piece.x, piece.y);
			ctx!.rotate(piece.rotation);

			ctx!.beginPath();
			ctx!.rect(-piece.size / 2, -piece.size / 4, piece.size, piece.size / 2);
			ctx!.fillStyle = piece.color + Math.floor(piece.opacity * config.opacity * 255).toString(16).padStart(2, '0');
			ctx!.fill();

			ctx!.restore();

			// Update
			piece.y += piece.speedY * speed;
			piece.x += piece.speedX * speed;
			piece.rotation += piece.rotationSpeed * speed;
			piece.speedX += (Math.random() - 0.5) * 0.1;

			// Reset if off screen
			if (piece.y > canvas.height + piece.size) {
				piece.y = -piece.size;
				piece.x = Math.random() * canvas.width;
			}
		});
	}

	function drawPulse(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		pulsePhase += 0.02 * speed;
		const pulse = (Math.sin(pulsePhase) + 1) / 2;

		const gradient = ctx.createRadialGradient(
			canvas.width / 2, canvas.height / 2, 0,
			canvas.width / 2, canvas.height / 2, Math.max(canvas.width, canvas.height) * 0.7
		);

		const color = colors[0] || '#667eea';
		gradient.addColorStop(0, color + Math.floor(pulse * config.opacity * 100).toString(16).padStart(2, '0'));
		gradient.addColorStop(0.5, color + Math.floor(pulse * config.opacity * 50).toString(16).padStart(2, '0'));
		gradient.addColorStop(1, color + '00');

		ctx.fillStyle = gradient;
		ctx.fillRect(0, 0, canvas.width, canvas.height);
	}

	function drawRipples(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		ripples.forEach((ripple, index) => {
			const fadeRatio = 1 - (ripple.radius / ripple.maxRadius);
			const alpha = ripple.opacity * fadeRatio * config.opacity;

			ctx!.beginPath();
			ctx!.arc(ripple.x, ripple.y, ripple.radius, 0, Math.PI * 2);
			ctx!.strokeStyle = ripple.color + Math.floor(alpha * 255).toString(16).padStart(2, '0');
			ctx!.lineWidth = 2;
			ctx!.stroke();

			// Expand ripple
			ripple.radius += ripple.speed * speed;

			// Reset if max radius reached
			if (ripple.radius >= ripple.maxRadius) {
				ripple.radius = 0;
				ripple.x = Math.random() * canvas.width;
				ripple.y = Math.random() * canvas.height;
				ripple.maxRadius = Math.random() * 100 + 50;
			}
		});
	}

	function drawNebula(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		nebulaTime += 0.005 * speed;

		// Draw nebula clouds
		for (let i = 0; i < 3; i++) {
			const x = canvas.width / 2 + Math.sin(nebulaTime + i * 2) * canvas.width * 0.3;
			const y = canvas.height / 2 + Math.cos(nebulaTime * 0.7 + i * 2) * canvas.height * 0.3;
			const size = Math.max(canvas.width, canvas.height) * (0.3 + i * 0.1);

			const gradient = ctx.createRadialGradient(x, y, 0, x, y, size);
			const color = colors[i % colors.length] || '#667eea';
			gradient.addColorStop(0, color + Math.floor(config.opacity * 60).toString(16).padStart(2, '0'));
			gradient.addColorStop(0.5, color + Math.floor(config.opacity * 30).toString(16).padStart(2, '0'));
			gradient.addColorStop(1, color + '00');

			ctx.fillStyle = gradient;
			ctx.fillRect(0, 0, canvas.width, canvas.height);
		}

		// Add some stars
		for (let i = 0; i < 50; i++) {
			const starX = (Math.sin(i * 12.3 + nebulaTime * 0.1) + 1) / 2 * canvas.width;
			const starY = (Math.cos(i * 7.7 + nebulaTime * 0.1) + 1) / 2 * canvas.height;
			const starSize = Math.sin(i * 3.3 + time * 0.001) * 0.5 + 1;

			ctx.beginPath();
			ctx.arc(starX, starY, starSize, 0, Math.PI * 2);
			ctx.fillStyle = `rgba(255, 255, 255, ${config.opacity * 0.5})`;
			ctx.fill();
		}
	}

	function drawCircuit(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		const color = colors[0] || '#3b82f6';

		// Draw connections
		circuitNodes.forEach((node, index) => {
			node.connections.forEach(targetIndex => {
				const target = circuitNodes[targetIndex];
				const pulse = (Math.sin(time * 0.002 + node.pulsePhase) + 1) / 2;

				ctx!.beginPath();
				ctx!.moveTo(node.x, node.y);
				ctx!.lineTo(target.x, target.y);
				ctx!.strokeStyle = color + Math.floor(pulse * config.opacity * 100).toString(16).padStart(2, '0');
				ctx!.lineWidth = 1;
				ctx!.stroke();
			});
		});

		// Draw nodes
		circuitNodes.forEach(node => {
			const pulse = (Math.sin(time * 0.002 + node.pulsePhase) + 1) / 2;

			ctx!.beginPath();
			ctx!.arc(node.x, node.y, 4, 0, Math.PI * 2);
			ctx!.fillStyle = color + Math.floor(pulse * config.opacity * 200).toString(16).padStart(2, '0');
			ctx!.fill();

			// Glow
			const gradient = ctx!.createRadialGradient(node.x, node.y, 0, node.x, node.y, 15);
			gradient.addColorStop(0, color + Math.floor(pulse * config.opacity * 80).toString(16).padStart(2, '0'));
			gradient.addColorStop(1, color + '00');
			ctx!.beginPath();
			ctx!.arc(node.x, node.y, 15, 0, Math.PI * 2);
			ctx!.fillStyle = gradient;
			ctx!.fill();

			node.pulsePhase += node.pulseSpeed * speed;
		});
	}

	function drawDots(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];
		const color = colors[0] || '#667eea';

		dots.forEach(dot => {
			dot.phase += dot.speed * speed;
			const pulse = (Math.sin(dot.phase) + 1) / 2;
			const size = dot.baseSize + pulse * 2;
			const alpha = 0.3 + pulse * 0.4;

			ctx!.beginPath();
			ctx!.arc(dot.x, dot.y, size, 0, Math.PI * 2);
			ctx!.fillStyle = color + Math.floor(alpha * config.opacity * 255).toString(16).padStart(2, '0');
			ctx!.fill();
		});
	}

	function drawFloatingShapes(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];

		floatingShapes.forEach(shape => {
			ctx!.save();
			ctx!.translate(shape.x, shape.y);
			ctx!.rotate(shape.rotation);

			ctx!.beginPath();
			if (shape.type === 'square') {
				ctx!.rect(-shape.size / 2, -shape.size / 2, shape.size, shape.size);
			} else if (shape.type === 'triangle') {
				ctx!.moveTo(0, -shape.size / 2);
				ctx!.lineTo(shape.size / 2, shape.size / 2);
				ctx!.lineTo(-shape.size / 2, shape.size / 2);
				ctx!.closePath();
			} else {
				ctx!.arc(0, 0, shape.size / 2, 0, Math.PI * 2);
			}

			ctx!.fillStyle = shape.color + Math.floor(shape.opacity * config.opacity * 100).toString(16).padStart(2, '0');
			ctx!.fill();
			ctx!.strokeStyle = shape.color + Math.floor(shape.opacity * config.opacity * 200).toString(16).padStart(2, '0');
			ctx!.lineWidth = 1;
			ctx!.stroke();

			ctx!.restore();

			// Update position and rotation
			shape.x += shape.speedX * speed;
			shape.y += shape.speedY * speed;
			shape.rotation += shape.rotSpeed * speed;

			// Wrap around
			if (shape.x < -shape.size) shape.x = canvas.width + shape.size;
			if (shape.x > canvas.width + shape.size) shape.x = -shape.size;
			if (shape.y < -shape.size) shape.y = canvas.height + shape.size;
			if (shape.y > canvas.height + shape.size) shape.y = -shape.size;
		});
	}

	function drawSmoke(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];
		const color = colors[0] || '#888888';

		smokeParticles.forEach(particle => {
			const gradient = ctx!.createRadialGradient(
				particle.x, particle.y, 0,
				particle.x, particle.y, particle.size
			);
			const alpha = particle.opacity * particle.life * config.opacity;
			gradient.addColorStop(0, color + Math.floor(alpha * 150).toString(16).padStart(2, '0'));
			gradient.addColorStop(0.5, color + Math.floor(alpha * 80).toString(16).padStart(2, '0'));
			gradient.addColorStop(1, color + '00');

			ctx!.beginPath();
			ctx!.arc(particle.x, particle.y, particle.size, 0, Math.PI * 2);
			ctx!.fillStyle = gradient;
			ctx!.fill();

			// Update
			particle.x += particle.speedX * speed;
			particle.y += particle.speedY * speed;
			particle.size += 0.3 * speed;
			particle.life -= 0.005 * speed;

			// Reset if faded
			if (particle.life <= 0) {
				particle.x = Math.random() * canvas.width;
				particle.y = canvas.height + 20;
				particle.size = Math.random() * 50 + 30;
				particle.opacity = Math.random() * 0.2 + 0.1;
				particle.life = 1;
			}
		});
	}

	function drawScanlines(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];
		const color = colors[0] || '#00ff00';

		scanlineOffset = (scanlineOffset + speed * 0.5) % 4;

		// Draw horizontal scanlines
		ctx.strokeStyle = color + Math.floor(config.opacity * 30).toString(16).padStart(2, '0');
		ctx.lineWidth = 1;

		for (let y = scanlineOffset; y < canvas.height; y += 4) {
			ctx.beginPath();
			ctx.moveTo(0, y);
			ctx.lineTo(canvas.width, y);
			ctx.stroke();
		}

		// Draw moving bright scanline
		const brightY = (time * 0.1 * speed) % canvas.height;
		const brightGradient = ctx.createLinearGradient(0, brightY - 20, 0, brightY + 20);
		brightGradient.addColorStop(0, color + '00');
		brightGradient.addColorStop(0.5, color + Math.floor(config.opacity * 100).toString(16).padStart(2, '0'));
		brightGradient.addColorStop(1, color + '00');

		ctx.fillStyle = brightGradient;
		ctx.fillRect(0, brightY - 20, canvas.width, 40);
	}

	function drawGrid(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];
		const color = colors[0] || '#3b82f6';

		// Draw horizontal lines
		gridLinesH.forEach(line => {
			const pulse = (Math.sin(time * 0.001 + line.pos * 0.01) + 1) / 2;
			ctx!.beginPath();
			ctx!.moveTo(0, line.pos);
			ctx!.lineTo(canvas.width, line.pos);
			ctx!.strokeStyle = color + Math.floor((line.opacity + pulse * 0.2) * config.opacity * 150).toString(16).padStart(2, '0');
			ctx!.lineWidth = 1;
			ctx!.stroke();
		});

		// Draw vertical lines
		gridLinesV.forEach(line => {
			const pulse = (Math.sin(time * 0.001 + line.pos * 0.01) + 1) / 2;
			ctx!.beginPath();
			ctx!.moveTo(line.pos, 0);
			ctx!.lineTo(line.pos, canvas.height);
			ctx!.strokeStyle = color + Math.floor((line.opacity + pulse * 0.2) * config.opacity * 150).toString(16).padStart(2, '0');
			ctx!.lineWidth = 1;
			ctx!.stroke();
		});

		// Draw intersection points
		gridLinesH.forEach(hLine => {
			gridLinesV.forEach(vLine => {
				const pulse = (Math.sin(time * 0.002 + hLine.pos * 0.01 + vLine.pos * 0.01) + 1) / 2;
				if (pulse > 0.7) {
					ctx!.beginPath();
					ctx!.arc(vLine.pos, hLine.pos, 3, 0, Math.PI * 2);
					ctx!.fillStyle = color + Math.floor(pulse * config.opacity * 255).toString(16).padStart(2, '0');
					ctx!.fill();
				}
			});
		});
	}

	function drawWarp(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];
		const centerX = canvas.width / 2;
		const centerY = canvas.height / 2;

		warpStars.forEach(star => {
			// Store previous position
			star.prevX = (star.x / star.z) * 200 + centerX;
			star.prevY = (star.y / star.z) * 200 + centerY;

			// Move star closer
			star.z -= speed * 10;

			// Calculate screen position
			const sx = (star.x / star.z) * 200 + centerX;
			const sy = (star.y / star.z) * 200 + centerY;

			// Reset if too close or off screen
			if (star.z <= 0 || sx < 0 || sx > canvas.width || sy < 0 || sy > canvas.height) {
				star.x = Math.random() * canvas.width - centerX;
				star.y = Math.random() * canvas.height - centerY;
				star.z = 1000;
				star.prevX = sx;
				star.prevY = sy;
				return;
			}

			// Draw line from previous to current (streak effect)
			const brightness = (1 - star.z / 1000) * config.opacity;
			ctx!.beginPath();
			ctx!.moveTo(star.prevX, star.prevY);
			ctx!.lineTo(sx, sy);
			ctx!.strokeStyle = `rgba(255, 255, 255, ${brightness})`;
			ctx!.lineWidth = (1 - star.z / 1000) * 3;
			ctx!.stroke();

			// Draw star point
			ctx!.beginPath();
			ctx!.arc(sx, sy, (1 - star.z / 1000) * 2, 0, Math.PI * 2);
			ctx!.fillStyle = `rgba(255, 255, 255, ${brightness})`;
			ctx!.fill();
		});
	}

	function drawHexgrid(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];
		const color = colors[0] || '#667eea';
		const hexSize = 30;

		hexCells.forEach(hex => {
			hex.phase += hex.speed * speed;
			const pulse = (Math.sin(hex.phase) + 1) / 2;

			ctx!.beginPath();
			for (let i = 0; i < 6; i++) {
				const angle = (i / 6) * Math.PI * 2 - Math.PI / 6;
				const x = hex.x + Math.cos(angle) * hexSize;
				const y = hex.y + Math.sin(angle) * hexSize;
				if (i === 0) {
					ctx!.moveTo(x, y);
				} else {
					ctx!.lineTo(x, y);
				}
			}
			ctx!.closePath();

			ctx!.strokeStyle = color + Math.floor((0.2 + pulse * 0.3) * config.opacity * 255).toString(16).padStart(2, '0');
			ctx!.lineWidth = 1;
			ctx!.stroke();

			// Fill on high pulse
			if (pulse > 0.7) {
				ctx!.fillStyle = color + Math.floor((pulse - 0.7) * config.opacity * 100).toString(16).padStart(2, '0');
				ctx!.fill();
			}
		});
	}

	function drawBinary(time: number) {
		if (!ctx) return;
		const config = intensityConfig[intensity];
		const color = colors[0] || '#00ff00';

		const fontSize = 12;
		ctx.font = `${fontSize}px 'Courier New', monospace`;

		// Draw falling binary
		matrixDrops.forEach(drop => {
			for (let i = 0; i < drop.chars.length; i++) {
				const y = drop.y - i * fontSize;
				if (y < 0 || y > canvas.height + fontSize) continue;

				const fadeRatio = 1 - (i / drop.chars.length);
				const alpha = fadeRatio * drop.opacity * config.opacity;

				// Use 0s and 1s instead of matrix chars
				const char = Math.random() > 0.5 ? '1' : '0';

				if (i === 0) {
					ctx!.fillStyle = `rgba(200, 255, 200, ${alpha})`;
				} else {
					ctx!.fillStyle = color + Math.floor(alpha * 200).toString(16).padStart(2, '0');
				}

				ctx!.fillText(char, drop.x, y);
			}

			drop.y += drop.speed * speed * 2;

			if (drop.y - drop.length * fontSize > canvas.height) {
				drop.y = -drop.length * fontSize;
				drop.x = Math.floor(Math.random() * (canvas.width / 15)) * 15 + 7;
			}
		});
	}

	function animate(time: number) {
		if (!ctx || effectType === 'none') return;

		// Don't clear for matrix (it has its own fade)
		if (effectType !== 'matrix') {
			ctx.clearRect(0, 0, canvas.width, canvas.height);
		}

		switch (effectType) {
			case 'particles':
				drawParticles();
				break;
			case 'gradient':
				drawGradient(time);
				break;
			case 'aurora':
				drawAurora(time);
				break;
			case 'starfield':
				drawStarfield(time);
				break;
			case 'matrix':
				drawMatrix(time);
				break;
			case 'waves':
				drawWaves(time);
				break;
			case 'bubbles':
				drawBubbles(time);
				break;
			case 'geometric':
				drawGeometric(time);
				break;
			case 'fireflies':
				drawFireflies(time);
				break;
			case 'rain':
				drawRain(time);
				break;
			case 'snow':
				drawSnow(time);
				break;
			case 'confetti':
				drawConfetti(time);
				break;
			case 'pulse':
				drawPulse(time);
				break;
			case 'ripples':
				drawRipples(time);
				break;
			case 'nebula':
				drawNebula(time);
				break;
			case 'circuit':
				drawCircuit(time);
				break;
			case 'dots':
				drawDots(time);
				break;
			case 'floatingShapes':
				drawFloatingShapes(time);
				break;
			case 'smoke':
				drawSmoke(time);
				break;
			case 'scanlines':
				drawScanlines(time);
				break;
			case 'grid':
				drawGrid(time);
				break;
			case 'warp':
				drawWarp(time);
				break;
			case 'hexgrid':
				drawHexgrid(time);
				break;
			case 'binary':
				drawBinary(time);
				break;
		}

		animationId = requestAnimationFrame(animate);
	}

	onMount(() => {
		if (!browser) return;

		initCanvas();
		window.addEventListener('resize', resizeCanvas);

		if (effectType !== 'none') {
			animationId = requestAnimationFrame(animate);
		}
	});

	onDestroy(() => {
		if (browser) {
			cancelAnimationFrame(animationId);
			window.removeEventListener('resize', resizeCanvas);
		}
	});

	// React to prop changes
	$effect(() => {
		if (browser && canvas) {
			cancelAnimationFrame(animationId);
			// Clear canvas for matrix effect transition
			if (ctx) {
				ctx.clearRect(0, 0, canvas.width, canvas.height);
			}
			initEffect();
			if (effectType !== 'none') {
				animationId = requestAnimationFrame(animate);
			}
		}
	});
</script>

{#if effectType !== 'none'}
	<canvas
		bind:this={canvas}
		class="animated-background"
		class:matrix-bg={effectType === 'matrix'}
	></canvas>
{/if}

<style>
	.animated-background {
		position: fixed;
		top: 0;
		left: 0;
		width: 100%;
		height: 100%;
		pointer-events: none;
		z-index: 0;
	}

	.matrix-bg {
		background: rgba(0, 0, 0, 0.9);
	}
</style>
