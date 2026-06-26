<script lang="ts">
	import { goto } from '$app/navigation';
	import { getSSEContext } from '$lib/sse.svelte';
	import { onMount } from 'svelte';
	import { HomeIcon } from '@lucide/svelte';

	const sse = getSSEContext();

	function returnToFrame() {
		goto('/kiosk');
	}

	// Auto-return: if dashboard_timeout_secs > 0, navigate back after that duration.
	onMount(() => {
		const timeoutSecs = sse.kiosk?.dashboard_timeout_secs ?? 0;
		if (timeoutSecs <= 0) return;
		const timer = setTimeout(returnToFrame, timeoutSecs * 1000);
		return () => clearTimeout(timer);
	});

	// If SSE is ready but no proxy URL is available, return to the frame.
	$effect(() => {
		if (sse.ready && !sse.kiosk?.dashboard_proxy_url) {
			goto('/kiosk');
		}
	});
</script>

{#if sse.kiosk?.dashboard_proxy_url}
	<div class="relative h-screen w-screen overflow-hidden">
		<!-- The proxy at localhost:8125 strips X-Frame-Options and CSP frame-ancestors
		     from HA's responses, making this iframe embedding work. -->
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
{:else}
	<div class="h-screen w-screen bg-black"></div>
{/if}
