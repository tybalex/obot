import { User } from "~/lib/model/users";
import { ApiRoutes, revalidateWhere } from "~/lib/routers/apiRoutes";
import { request } from "~/lib/service/api/primitives";

async function getMe() {
    const res = await request<User>({
        url: ApiRoutes.me().url,
        errorMessage: "Failed to fetch agents",
    });

    return res.data;
}
getMe.key = () => ({ url: ApiRoutes.me().path }) as const;

const revalidateMe = () =>
    revalidateWhere((url) => url.includes(ApiRoutes.me().path));

export const UserService = {
    getMe,
    revalidateMe,
};
