<script lang="ts">
	/**
	 * BoeingAITokenPrompt - Shows a banner/prompt when user hasn't configured their UDAL PAT.
	 * Place this in layouts where Boeing AI models are used (e.g., agent chat).
	 * Only shows if the user hasn't stored their credential yet.
	 */

	import { getBoeingAICredentialStatus } from '$lib/services/user/boeing-ai';
	import { onMount } from 'svelte';

	let needsSetup = $state(false);
	let dismissed = $state(false);
	let showSetupDialog = $state(false);

	onMount(async () => {
		const status = await getBoeingAICredentialStatus();
		needsSetup = !status.configured;
	});
</script>

{#if needsSetup && !dismissed}
	<div
		class="bg-warning/10 border-warning/30 flex items-center justify-between gap-3 rounded-lg border px-4 py-3"
	>
		<div class="flex items-center gap-3">
			<span class="text-warning text-lg">⚠️</span>
			<div>
				<p class="text-sm font-medium">Boeing AI not configured</p>
				<p class="text-base-content/60 text-xs">
					Enter your UDAL Personal Access Token to use Boeing AI models.
				</p>
			</div>
		</div>
		<div class="flex gap-2">
			<button class="btn btn-xs btn-ghost" onclick={() => (dismissed = true)}> Later </button>
			<a href="/keys" class="btn btn-xs btn-primary"> Configure </a>
		</div>
	</div>
{/if}
