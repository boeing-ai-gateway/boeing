import { handleRouteError } from '$lib/errors';
import { UserService, BoeingbotService, type OrgUser } from '$lib/services';
import type { ProjectV2Agent } from '$lib/services/boeingbot/types';
import type { PageLoad } from './$types';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ fetch, parent }) => {
	const { profile, version } = await parent();

	if (version?.agentsEnabled === false) {
		throw error(403, 'Boeing Agent features are disabled.');
	}

	let agents: ProjectV2Agent[] = [];
	let users: OrgUser[] = [];
	try {
		[agents, users] = await Promise.all([
			BoeingbotService.listAllBoeingbotAgents({ fetch }),
			UserService.listUsers({ fetch })
		]);
	} catch (err) {
		handleRouteError(err, `/agents`, profile);
	}

	return {
		agents,
		users
	};
};
