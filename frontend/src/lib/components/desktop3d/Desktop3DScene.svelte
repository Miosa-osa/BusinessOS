<script lang="ts">
	import { T } from '@threlte/core';
	import { OrbitControls, interactivity } from '@threlte/extras';
	import * as THREE from 'three';
	import Desktop3DWindow from './Desktop3DWindow.svelte';
	import type { Window3DState, ViewMode } from '$lib/stores/desktop3dStore';

	// Enable pointer events on 3D objects
	interactivity();

	interface Props {
		windows: Window3DState[];
		viewMode: ViewMode;
		focusedWindowId: string | null;
		autoRotate: boolean;
		sphereRadius: number;
		onWindowClick?: (id: string) => void;
		onBackgroundClick?: () => void;
		onResize?: (widthDelta: number, heightDelta: number) => void;
		onZoomOut?: () => void; // Called when user zooms out in focus mode
	}

	let {
		windows = [],
		viewMode = 'orb',
		focusedWindowId = null,
		autoRotate = true,
		sphereRadius = 60,
		onWindowClick,
		onBackgroundClick,
		onResize,
		onZoomOut
	}: Props = $props();

	// Track camera distance for zoom-out detection
	let orbitControlsRef: any = $state(null);

	// Track which window is hovered (only ONE at a time)
	let hoveredWindowId: string | null = $state(null);

	// Handle zoom out while focused - exit focus mode
	// Only trigger on significant zoom out, not rotation
	function handleControlsChange(e: any) {
		if (!focusedWindowId || !orbitControlsRef) return;

		// Get current camera distance from target
		const controls = orbitControlsRef;
		if (controls?.object) {
			const distance = controls.object.position.distanceTo(controls.target);
			// Only unfocus if zoomed WAY out (past the orb view distance)
			if (distance > 600) {
				onZoomOut?.();
			}
		}
	}
	// NO camera reset - keep free-form rotation at all times

	// Initial camera position - OrbitControls will manage from here
	// No spring - let user freely rotate/zoom
	// Backed up to accommodate larger sphere radius (95)
	const initialCameraPosition: [number, number, number] = [0, 40, 300];

	// Effective auto-rotate (disabled when focused)
	let effectiveAutoRotate = $derived(autoRotate && !focusedWindowId);

	// Reset camera to front view when focusing a window
	$effect(() => {
		if (focusedWindowId && orbitControlsRef) {
			// Smoothly move camera to face the focused window (which is at z=200)
			const controls = orbitControlsRef;
			if (controls?.object) {
				// Reset camera to front position
				controls.object.position.set(0, 40, 350);
				controls.target.set(0, 0, 0);
				controls.update();
			}
		}
	});

	// Calculate indices for prev/next windows
	let focusedIndex = $derived(windows.findIndex(w => w.id === focusedWindowId));
	let prevIndex = $derived(focusedIndex > 0 ? focusedIndex - 1 : windows.length - 1);
	let nextIndex = $derived(focusedIndex >= 0 && focusedIndex < windows.length - 1 ? focusedIndex + 1 : 0);

	// Handle background click
	function handleBackgroundClick() {
		onBackgroundClick?.();
	}

</script>

<!-- Camera - OrbitControls manages position freely -->
<T.PerspectiveCamera
	makeDefault
	position={initialCameraPosition}
	fov={50}
>
	<OrbitControls
		bind:ref={orbitControlsRef}
		enableDamping={true}
		dampingFactor={0.08}
		autoRotate={effectiveAutoRotate}
		autoRotateSpeed={0.3}
		minDistance={150}
		maxDistance={800}
		minPolarAngle={0.1}
		maxPolarAngle={Math.PI * 0.9}
		enablePan={false}
		enableZoom={true}
		enableRotate={true}
		onchange={handleControlsChange}
	/>
</T.PerspectiveCamera>

<!-- Lighting -->
<T.AmbientLight intensity={0.6} />
<T.DirectionalLight position={[50, 100, 50]} intensity={1} />
<T.DirectionalLight position={[-50, 50, -50]} intensity={0.5} />
<T.PointLight position={[0, 0, 0]} intensity={2} color="#4a9eff" distance={100} />

<!-- Background click catcher -->
<T.Mesh
	onclick={handleBackgroundClick}
>
	<T.SphereGeometry args={[500, 32, 32]} />
	<T.MeshBasicMaterial
		visible={false}
		side={THREE.BackSide}
	/>
</T.Mesh>

<!-- NO center orb - windows are the focus -->
<!-- NO connection lines -->

<!-- Module Windows -->
{#each windows as window, index (window.id)}
	<Desktop3DWindow
		{window}
		isFocused={window.id === focusedWindowId}
		isPrevWindow={focusedWindowId !== null && index === prevIndex}
		isNextWindow={focusedWindowId !== null && index === nextIndex}
		isHovered={window.id === hoveredWindowId}
		viewMode={viewMode}
		onClick={() => onWindowClick?.(window.id)}
		onResize={onResize}
		onHover={(hovered) => {
			if (hovered) {
				hoveredWindowId = window.id;
			} else if (hoveredWindowId === window.id) {
				hoveredWindowId = null;
			}
		}}
	/>
{/each}

<!-- Background handled by parent container gradient -->
