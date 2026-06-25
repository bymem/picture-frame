<script lang="ts">
	import { getSSEContext } from '$lib/sse.svelte';

	const sse = getSSEContext();

	// Navigate to the dashboard URL as soon as it's available from SSE.
	// Full-page navigation avoids X-Frame-Options / CSP blocks from HA.
	$effect(() => {
		const url = sse.kiosk?.dashboard_url;
		if (url) {
			window.location.replace(url);
		} else if (sse.ready) {
			// No URL configured — go back to the frame.
			window.location.replace('/kiosk');
		}
	});
</script>

<!-- Visible only for the fraction of a second before SSE delivers the URL. -->
<div class="h-screen w-screen bg-black"></div>
