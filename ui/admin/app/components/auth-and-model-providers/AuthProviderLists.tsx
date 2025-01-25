import { CircleCheckIcon, CircleHelpIcon, CircleSlashIcon } from "lucide-react";
import { Link } from "react-router";

import { AuthProvider } from "~/lib/model/providers";

import { ProviderConfigure } from "~/components/auth-and-model-providers/ProviderConfigure";
import { ProviderIcon } from "~/components/auth-and-model-providers/ProviderIcon";
import { ProviderMenu } from "~/components/auth-and-model-providers/ProviderMenu";
import { AuthProviderLinks } from "~/components/auth-and-model-providers/constants";
import { Badge } from "~/components/ui/badge";
import { Button } from "~/components/ui/button";
import { Card, CardContent, CardHeader } from "~/components/ui/card";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "~/components/ui/tooltip";

function isDisabled(
	authProvider: AuthProvider,
	authProviders: AuthProvider[]
): boolean {
	return (
		!authProvider.configured &&
		authProviders.filter((p) => p.configured).length > 0
	);
}

export function AuthProviderList({
	authProviders,
}: {
	authProviders: AuthProvider[];
}) {
	return (
		<div className="space-y-4">
			<div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 2xl:grid-cols-4">
				{authProviders.map((authProvider) => (
					<Card key={authProvider.id}>
						<CardHeader className="flex flex-row items-center justify-between pb-4 pt-2">
							{authProvider.configured && (
								<div className="flex flex-row items-center gap-2">
									<ProviderMenu provider={authProvider} />
								</div>
							)}
							{isDisabled(authProvider, authProviders) && (
								<div className="flex flex-row items-center gap-2">
									<Tooltip>
										<TooltipTrigger asChild>
											<Button variant="ghost" size="icon">
												<CircleHelpIcon />
											</Button>
										</TooltipTrigger>
										<TooltipContent>
											<span>
												You must deconfigure the existing provider before
												configuring this one.
											</span>
										</TooltipContent>
									</Tooltip>
								</div>
							)}
							{!isDisabled(authProvider, authProviders) &&
								!authProvider.configured && (
									<div className="flex flex-row items-center gap-2" />
								)}
						</CardHeader>
						<CardContent className="flex flex-col items-center gap-4">
							<Link to={AuthProviderLinks[authProvider.id]}>
								<ProviderIcon provider={authProvider} size="lg" />
							</Link>
							<div className="text-center text-lg font-semibold">
								{authProvider.name}
							</div>

							<Badge className="pointer-events-none" variant="outline">
								{authProvider.configured ? (
									<span className="flex items-center gap-1">
										<CircleCheckIcon size={18} className="text-success" />{" "}
										Configured
									</span>
								) : (
									<span className="flex items-center gap-1">
										<CircleSlashIcon size={18} className="text-destructive" />
										Not Configured
									</span>
								)}
							</Badge>
							<ProviderConfigure
								provider={authProvider}
								disabled={isDisabled(authProvider, authProviders)}
							/>
						</CardContent>
					</Card>
				))}
			</div>
		</div>
	);
}
