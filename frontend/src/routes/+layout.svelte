<script lang="ts">
	import '../app.css';
	let { children } = $props();
</script>

<svelte:head>
	<title>Business OS</title>
	<meta name="description" content="Your internal command center" />
</svelte:head>

{@render children()}
