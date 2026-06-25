<script lang="ts">
	import { goto } from '$app/navigation';
	import { getSSEContext } from '$lib/sse.svelte';
	import { XIcon } from '@lucide/svelte';

	const sse = getSSEContext();

	let dashboardUrl = $derived(sse.kiosk?.dashboard_url ?? '');
	let timeoutSecs = $derived(sse.kiosk?.dashboard_timeout_secs ?? 0);

	function returnToFrame() {
		goto('/kiosk');
	}

	// If SSE is ready but no dashboard URL is configured, go back immediately.
	$effect(() => {
		if (sse.ready && !dashboardUrl) {
			returnToFrame();
		}
	});

	// Auto-return timer — runs whenever timeoutSecs becomes available from SSE.
	// Pointer events (touch or mouse) reset the idle window.
	$effect(() => {
		if (timeoutSecs <= 0) return;

		let timer: ReturnType<typeof setTimeout> | null = null;

		function resetTimer() {
			if (timer !== null) clearTimeout(timer);
			timer = setTimeout(returnToFrame, timeoutSecs * 1000);
		}

		resetTimer();
		window.addEventListener('pointerdown', resetTimer);

		return () => {
			if (timer !== null) clearTimeout(timer);
			window.removeEventListener('pointerdown', resetTimer);
		};
	});
</script>

<div class="relative h-screen w-screen overflow-hidden bg-black">
	{#if dashboardUrl}
		<iframe
			title="Dashboard"
			src={dashboardUrl}
			class="h-full w-full border-none"
			allow="fullscreen"
		></iframe>
	{/if}

	<button
		type="button"
		onclick={returnToFrame}
		class="absolute right-4 bottom-4 flex items-center gap-2 rounded-full bg-black/60 px-4 py-2.5 text-sm font-medium text-white backdrop-blur-sm transition-opacity hover:bg-black/80"
		aria-label="Return to picture frame"
		data-testid="dashboard-back-button"
	>
		<XIcon class="size-4" />
		Back to frame
	</button>
</div>
