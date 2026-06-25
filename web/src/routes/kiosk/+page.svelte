<script lang="ts">
	import { goto } from '$app/navigation';
	import { Heartbeat } from '$lib/heartbeat';
	import { getSSEContext } from '$lib/sse.svelte';
	import { reloadOnBackendVersionChange } from '$lib/versionReload.svelte';
	import { onMount } from 'svelte';
	import Images from './components/Images.svelte';
	import Overlay from './components/Overlay.svelte';

	const sse = getSSEContext();

	onMount(() => {
		const heartbeat = new Heartbeat();
		heartbeat.start();
		return () => heartbeat.stop();
	});

	// Reload onto the new bundle after a self-update swaps the binary.
	reloadOnBackendVersionChange(() => sse.kiosk?.version);

	function handleTouch() {
		if (!sse.kiosk?.dashboard_url) return;
		goto('/kiosk/dashboard');
	}
</script>

{#if sse.ready}
	<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
	<div class="h-screen w-screen overflow-hidden" onclick={handleTouch}>
		<Images />
		<Overlay />
	</div>
{:else}
	<div class="h-screen w-screen overflow-hidden bg-black"></div>
{/if}
