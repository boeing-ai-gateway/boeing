import { BoeingbotService } from '$lib/services';
import type { PageLoad } from './$types';
import { redirect } from '@sveltejs/kit';

export const ssr = false;

export const load: PageLoad = async ({ fetch }) => {
	const projects = await BoeingbotService.listProjects({ fetch });
	if (projects.length === 0) {
		const project = await BoeingbotService.createProject({ displayName: 'New Project' }, { fetch });
		throw redirect(302, `/agent/p/${project.id}`);
	}
	throw redirect(302, `/agent/p/${projects[0].id}`);
};
