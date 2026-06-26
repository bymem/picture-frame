<script lang="ts">
	import { goto } from '$app/navigation';
	import { getSSEContext } from '$lib/sse.svelte';
	import { onMount } from 'svelte';
	import { HomeIcon } from '@lucide/svelte';

	const sse = getSSEContext();

	function returnToFrame() {
		goto('/kiosk');
	}

	// Auto-return timer: if dashboard_timeout_secs > 0, navigate back after that duration.
	// The cleanup function fires if the user leaves early (e.g. presses the back button).
	onMount(() => {
		const timeoutSecs = sse.kiosk?.dashboard_timeout_secs ?? 0;
		if (timeoutSecs <= 0) return;

		const timer = setTimeout(returnToFrame, timeoutSecs * 1000);
		return () => clearTimeout(timer);
	});

	// If SSE is ready but no URL is configured, return to the frame.
	$effect(() => {
		if (sse.ready && !sse.kiosk?.dashboard_url) {
			goto('/kiosk');
		}
	});
</script>

{#if sse.kiosk?.dashboard_url}
	<div class="relative h-screen w-screen overflow-hidden">
		<iframe
			src={sse.kiosk.dashboard_url}
			title="Home Assistant dashboard"
			class="h-full w-full border-0"
		></iframe>

		<!-- Floating return button, unobtrusive in the top-left corner -->
		<button
			onclick={returnToFrame}
			aria-label="Return to picture frame"
			class="absolute left-4 top-4 flex cursor-pointer items-center gap-1.5 rounded-full bg-black/50 px-3 py-1.5 text-xs text-white backdrop-blur-sm hover:bg-black/70"
		>
			<HomeIcon class="size-3.5" />
			Frame
		</button>
	</div>
{:else}
	<!-- Shown while SSE is loading before the URL is known. -->
	<div class="h-screen w-screen bg-black"></div>
{/if}
