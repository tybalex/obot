import { BoxesIcon } from "lucide-react";

import { ModelProvider } from "~/lib/model/modelProviders";
import { cn } from "~/lib/utils";

import { CommonModelProviderIds } from "~/components/model-providers/constants";

export function ModelProviderIcon({
	modelProvider,
	size = "md",
}: {
	modelProvider: ModelProvider;
	size?: "md" | "lg";
}) {
	const ignoreDarkModeSet = new Set([
		CommonModelProviderIds.AZURE_OPENAI,
		CommonModelProviderIds.DEEPSEEK,
	]);

	return modelProvider.icon ? (
		<img
			src={modelProvider.icon}
			alt={modelProvider.name}
			className={cn({
				"h-6 w-6": size === "md",
				"h-16 w-16": size === "lg",
				"dark:invert": !ignoreDarkModeSet.has(modelProvider.id),
			})}
		/>
	) : (
		<BoxesIcon className="color-primary h-16 w-16" />
	);
}
