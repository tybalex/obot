import { z } from "zod";

import {
	OAuthAppSpec,
	OAuthFormStep,
	getOAuthLinks,
} from "~/lib/model/oauthApps/oauth-helpers";
import { assetUrl } from "~/lib/utils";

const schema = z.object({
	clientID: z.string().min(1, "Client ID is required"),
	clientSecret: z.string().min(1, "Client Secret is required"),
});

const scopes = [
	"email",
	"identify",
	"guilds",
	"messages.read",
];

const steps: OAuthFormStep<z.infer<typeof schema>>[] = [
	{
		type: "markdown",
		text:
			"### Step 1: Create a new app in Zoom App Marketplace\n" +
			"If you already have an app, you can skip to Step 2.\n\n" +
			"- Ensure you are logged into your preferred Zoom account.\n" +
			"- From the [Zoom App Marketplace](https://marketplace.zoom.us/), hover over the **Develop** box on the top right corner and then select **Build App**.\n" +
			"- When asked `What kind of app are you creating?` in the pop-up window, select **General App** and click **Create**.\n" +
			"- Now You should be redirected to the **App Basic Information** page.\n" +
			"- In the **Select how the app is managed** section, make sure it is set to **User-managed**.\n" +
			"- Scroll down to the **Oauth Information** section, copy the url below and paste it into the **OAuth Redirect URL** field.\n",
	},
	{
		type: "copy",
		text: getOAuthLinks("discord").redirectURL,
	},
	{
		type: "markdown",
		text: "- Click **Continue** at the bottom and then click **Scopes** on the left sidebar to continue.\n",
	},
	{
		type: "markdown",
		text:
			"### Step 2: Configure the app's necessary scopes\n" +
			"- From the [Zoom App Management Page](https://marketplace.zoom.us/user/build), click on the app you created and click **Scopes** on the left sidebar.\n" +
			"(You will already be on this page if you've completed step 1)\n" +
			"- Click on **Add Scopes** to add necessary scopes for the zoom tool. Ensure the following scopes are added:\n",
	},
	{
		type: "sectionGroup",
		sections: [
			{
				title: "Scopes: ",
				displayStepsInline: true,
				defaultOpen: true,
				steps: scopes.map((scope) => ({
					type: "copy",
					text: scope,
				})),
			},
		],
	},
	{
		type: "markdown",
		text:
			"### Step 3: Register your App with Obot\n" +
			"- **Client ID** and **Client Secret** can be found in the **App Credentials** box, located at the top of the left sidebar. Alternatively, you can find them in the **App Credentials** section of the **Basic Information** page.\n" +
			"- Copy the **Client ID** and **Client Secret** and paste them into the respective fields below.\n",
	},
	{ type: "input", input: "clientID", label: "Client ID" },
	{
		type: "input",
		input: "clientSecret",
		label: "Client Secret",
		inputType: "password",
	},
];

export const DiscordOAuthApp = {
	schema,
	alias: "discord",
	type: "discord",
	displayName: "Discord",
	logo: assetUrl("/assets/discord-logo.svg"),
	steps: steps,
	noGatewayIntegration: true,
} satisfies OAuthAppSpec;
