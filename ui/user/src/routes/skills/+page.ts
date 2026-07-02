import { handleRouteError, HttpError } from '$lib/errors';
import { BoeingbotService } from '$lib/services';
import type { Skill } from '$lib/services/boeingbot/types';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ fetch, parent }) => {
	const { profile } = await parent();
	let skills: Skill[] = [];
	let showLicenseError = false;

	try {
		skills = await BoeingbotService.listSkills({ fetch, dontLogErrors: true });
	} catch (err) {
		if (err instanceof HttpError && err.statusCode === 402) {
			skills = [];
			showLicenseError = true;
		} else {
			handleRouteError(err, `/skills`, profile);
		}
	}

	return {
		skills,
		showLicenseError
	};
};
