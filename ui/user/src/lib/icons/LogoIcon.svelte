<script lang="ts">
	import { appPreferences } from '$lib/stores';

	let logoSrc = $derived(appPreferences.current.logos?.logoIcon);
</script>

<img
	class="size-5"
	role="img"
	alt="Boeing logo icon"
	src={logoSrc}
/>
