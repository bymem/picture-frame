<script lang="ts">
	import { Heartbeat } from '$lib/heartbeat';
	import { getSSEContext } from '$lib/sse.svelte';
	import { reloadOnBackendVersionChange } from '$lib/versionReload.svelte';
	import { onMount } from 'svelte';
	import { HomeIcon } from '@lucide/svelte';
	import Images from './components/Images.svelte';
	import Overlay from './components/Overlay.svelte';

	const sse = getSSEContext();
	let showDashboard = $state(false);
	let returnTimer: ReturnType<typeof setTimeout> | null = null;

	onMount(() => {
		const heartbeat = new Heartbeat();
		heartbeat.start();
		return () => heartbeat.stop();
	});

	reloadOnBackendVersionChange(() => sse.kiosk?.version);

	function openDashboard() {
		if (!sse.kiosk?.dashboard_proxy_url) return;
		showDashboard = true;
		scheduleReturn();
	}

	function returnToFrame() {
		showDashboard = false;
		clearReturnTimer();
	}

	function scheduleReturn() {
		clearReturnTimer();
		const secs = sse.kiosk?.dashboard_timeout_secs ?? 0;
		if (secs <= 0) return;
		returnTimer = setTimeout(returnToFrame, secs * 1000);
	}

	function clearReturnTimer() {
		if (returnTimer !== null) {
			clearTimeout(returnTimer);
			returnTimer = null;
		}
	}
</script>

{#if sse.ready}
	<!-- Frame layer — click/tap opens the dashboard -->
	<!-- svelte-ignore a11y_click_events_have_key_events a11y_no_static_element_interactions -->
	<div class="h-screen w-screen overflow-hidden" onclick={openDashboard}>
		<Images />
		<Overlay />
	</div>

	<!-- Dashboard layer: always rendered so HA loads in the background.
	     visibility:hidden keeps the iframe alive (vs display:none which may pause it). -->
	{#if sse.kiosk?.dashboard_proxy_url}
		<div
			class="fixed inset-0 z-10 transition-opacity duration-300"
			class:invisible={!showDashboard}
			class:pointer-events-none={!showDashboard}
			class:opacity-0={!showDashboard}
		>
			<iframe
				src={sse.kiosk.dashboard_proxy_url}
				title="Home Assistant dashboard"
				class="h-full w-full border-0"
			></iframe>

			<button
				onclick={returnToFrame}
				aria-label="Return to picture frame"
				class="absolute left-4 top-4 flex cursor-pointer items-center gap-1.5 rounded-full bg-black/50 px-3 py-1.5 text-xs text-white backdrop-blur-sm hover:bg-black/70"
			>
				<HomeIcon class="size-3.5" />
				Frame
			</button>
		</div>
	{/if}
{:else}
	<div class="h-screen w-screen overflow-hidden bg-black"></div>
{/if}
