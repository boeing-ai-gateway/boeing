<script lang="ts">
	import {
		getBoeingAICredentialStatus,
		saveBoeingAICredential,
		deleteBoeingAICredential,
		type BoeingAICredentialStatus
	} from '$lib/services/user/boeing-ai';
	import { onMount } from 'svelte';

	let status = $state<BoeingAICredentialStatus>({ configured: false });
	let tokenInput = $state('');
	let saving = $state(false);
	let error = $state('');
	let success = $state('');
	let showDialog = $state(false);
	let deleting = $state(false);

	const UDAL_DEVELOPERS_URL = 'https://udal.web.boeing.com/udaladmin/profile/developers';
	const UDAL_TEST_DEVELOPERS_URL =
		'https://udal-test.web.boeing.com/udaladmin/profile/developers';

	onMount(async () => {
		status = await getBoeingAICredentialStatus();
	});

	async function handleSave() {
		if (!tokenInput.trim()) {
			error = 'Please enter your UDAL Personal Access Token';
			return;
		}

		saving = true;
		error = '';
		success = '';

		const result = await saveBoeingAICredential(tokenInput.trim());

		if (result.success) {
			success = 'Boeing AI credential saved successfully';
			tokenInput = '';
			status = await getBoeingAICredentialStatus();
			setTimeout(() => {
				showDialog = false;
				success = '';
			}, 1500);
		} else {
			error = result.error || 'Failed to save credential';
		}

		saving = false;
	}

	async function handleDelete() {
		deleting = true;
		error = '';
		try {
			await deleteBoeingAICredential();
			status = { configured: false };
			success = 'Credential removed';
			setTimeout(() => (success = ''), 2000);
		} catch (e) {
			error = 'Failed to remove credential';
		}
		deleting = false;
	}

	function openUDALPage() {
		window.open(UDAL_TEST_DEVELOPERS_URL, '_blank', 'noopener,noreferrer');
	}
</script>

<!-- Token Status Card -->
<div class="rounded-box border-base-300 bg-base-100 border p-4">
	<div class="flex items-center justify-between">
		<div>
			<h3 class="text-base font-medium">Boeing AI (BCAI) Access</h3>
			<p class="text-base-content/60 mt-1 text-sm">
				{#if status.configured}
					<span class="text-success">● Connected</span>
					{#if status.maskedToken}
						— Token: {status.maskedToken}
					{/if}
					{#if status.expiresAt}
						<span class="text-base-content/40"> · Expires: {status.expiresAt}</span>
					{/if}
				{:else}
					<span class="text-warning">● Not configured</span> — Enter your UDAL PAT to use
					Boeing AI models
				{/if}
			</p>
		</div>
		<div class="flex gap-2">
			{#if status.configured}
				<button class="btn btn-sm btn-ghost" onclick={() => (showDialog = true)}>
					Update
				</button>
				<button
					class="btn btn-sm btn-error btn-outline"
					onclick={handleDelete}
					disabled={deleting}
				>
					{deleting ? 'Removing...' : 'Remove'}
				</button>
			{:else}
				<button class="btn btn-sm btn-primary" onclick={() => (showDialog = true)}>
					Configure
				</button>
			{/if}
		</div>
	</div>
</div>

<!-- Token Setup Dialog -->
{#if showDialog}
	<dialog class="modal modal-open" aria-modal="true">
		<div class="modal-box max-w-lg">
			<h3 class="text-lg font-bold">Configure Boeing AI Access</h3>

			<div class="mt-4 space-y-4">
				<!-- Instructions -->
				<div class="bg-info/10 rounded-box p-3 text-sm">
					<p class="font-medium">To get your UDAL Personal Access Token:</p>
					<ol class="mt-2 list-inside list-decimal space-y-1">
						<li>Open the UDAL Developers page (link below)</li>
						<li>Sign in with your Boeing credentials</li>
						<li>Click "Generate Personal Access Token" (or copy existing)</li>
						<li>Paste the token below</li>
					</ol>
					<p class="mt-2 text-xs opacity-70">
						The PAT is valid for 90 days and is unique to you. It enables per-user audit
						tracking of AI usage.
					</p>
				</div>

				<!-- Open UDAL Button -->
				<button class="btn btn-outline btn-sm w-full" onclick={openUDALPage}>
					Open UDAL Developers Page ↗
				</button>

				<!-- Token Input -->
				<div class="form-control">
					<label class="label" for="udal-pat-input">
						<span class="label-text">UDAL Personal Access Token</span>
					</label>
					<input
						id="udal-pat-input"
						type="password"
						class="input input-bordered w-full font-mono text-sm"
						placeholder="Paste your PAT here..."
						bind:value={tokenInput}
						onkeydown={(e) => e.key === 'Enter' && handleSave()}
					/>
				</div>

				<!-- Error/Success Messages -->
				{#if error}
					<div class="text-error text-sm">{error}</div>
				{/if}
				{#if success}
					<div class="text-success text-sm">{success}</div>
				{/if}
			</div>

			<!-- Actions -->
			<div class="modal-action">
				<button
					class="btn btn-ghost"
					onclick={() => {
						showDialog = false;
						error = '';
						tokenInput = '';
					}}
				>
					Cancel
				</button>
				<button class="btn btn-primary" onclick={handleSave} disabled={saving || !tokenInput}>
					{saving ? 'Validating...' : 'Save Token'}
				</button>
			</div>
		</div>
		<form method="dialog" class="modal-backdrop">
			<button
				onclick={() => {
					showDialog = false;
					error = '';
					tokenInput = '';
				}}>close</button
			>
		</form>
	</dialog>
{/if}
