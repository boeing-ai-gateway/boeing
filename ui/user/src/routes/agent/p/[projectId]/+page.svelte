<script lang="ts">
	import { page } from '$app/state';
	import ProjectStartThread from '$lib/components/boeingbot/ProjectStartThread.svelte';
	import type { ProjectLayoutContext } from '$lib/services/boeingbot/types';
	import { PROJECT_LAYOUT_CONTEXT } from '$lib/services/boeingbot/types';
	import { profile } from '$lib/stores';
	import { boeingbotChat } from '$lib/stores/boeingbotChat.svelte';
	import { getContext } from 'svelte';

	let { data } = $props();
	let agent = $derived(data.agent);
	let projectId = $derived(data.projects[0].id);
	let tid = $derived(page.url.searchParams.get('tid'));
	let session = $derived($boeingbotChat?.sessions?.find((s) => s.id === tid));
	let browserBaseUrl = $derived(data.agent.connectURL);

	const projectLayout = getContext<ProjectLayoutContext>(PROJECT_LAYOUT_CONTEXT);

	let displayChat = $derived($boeingbotChat?.chat);
	let impersonating = $derived(data.agent.userID !== profile.current.id);
</script>

{#if displayChat}
	{#key displayChat}
		<ProjectStartThread
			agentId={agent.id}
			{projectId}
			{browserBaseUrl}
			browserAvailable={projectLayout.browserAvailable}
			bind:browserViewerOpen={projectLayout.browserViewerOpen}
			chat={displayChat}
			onFileOpen={projectLayout.handleFileOpen}
			suppressEmptyState
			onThreadContentWidth={projectLayout.setThreadContentWidth}
			classes={{ root: impersonating ? 'h-[calc(100dvh-8rem)]' : 'h-[calc(100dvh-4rem)]' }}
		/>
	{/key}
{/if}

<svelte:head>
	<title>Boeing | {session?.title || 'Untitled'}</title>
</svelte:head>
