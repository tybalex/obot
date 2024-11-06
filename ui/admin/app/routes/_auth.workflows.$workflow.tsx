import { Editor } from "@monaco-editor/react";
import {
    ClientLoaderFunctionArgs,
    redirect,
    useLoaderData,
} from "@remix-run/react";
import { dump } from "js-yaml";
import { $params } from "remix-routes";

import { WorkflowService } from "~/lib/service/api/workflowService";
import { noop } from "~/lib/utils";

import { useTheme } from "~/components/theme";

export const clientLoader = async ({ params }: ClientLoaderFunctionArgs) => {
    const { workflow: id } = $params("/workflows/:workflow", params);

    if (!id) {
        throw redirect("/threads");
    }

    const workflow = await WorkflowService.getWorkflowById(id).catch(noop);
    if (!workflow) throw redirect("/agents");

    return { workflow };
};

export default function ChatAgent() {
    const { isDark } = useTheme();
    const { workflow } = useLoaderData<typeof clientLoader>();

    const disclaimer =
        "# This is a read-only view of this workflow.\n# To create, manage, and run, worfklows, please use the otto CLI";

    return (
        <div className="flex flex-col h-full">
            <div className="flex-1">
                <div className="h-full overflow-hidden">
                    <Editor
                        height="100%"
                        defaultLanguage="yaml"
                        defaultValue={
                            disclaimer +
                            "\n\n" +
                            dump(
                                JSON.parse(
                                    JSON.stringify(workflow, (_, v) =>
                                        v === null ? undefined : v
                                    )
                                )
                            )
                        }
                        // note:(tylerslaton): There is a big long process to get the theme to be different here. We'll want to do
                        // that when an overhaul to theming is done. However, this will work for now since we'll replace this very
                        // soon anyway.
                        theme={isDark ? "hc-black" : "vs"}
                        options={{
                            padding: { top: 20 },
                            minimap: { enabled: false },
                            readOnly: true,
                            lineNumbers: "off",
                            fontSize: 14,
                            fontFamily: "monospace",
                            wordWrap: "on",
                            automaticLayout: true,
                            scrollBeyondLastLine: false,
                            cursorBlinking: "solid",
                            cursorStyle: "underline",
                            renderLineHighlight: "none",
                            domReadOnly: true,
                            selectionHighlight: false,
                            occurrencesHighlight: "off",
                            guides: {
                                indentation: false,
                            },
                            scrollbar: {
                                vertical: "hidden",
                            },
                            matchBrackets: "never",
                            renderWhitespace: "none",
                            colorDecorators: false,
                            links: false,
                            hover: {
                                enabled: false,
                            },
                        }}
                    />
                </div>
            </div>
        </div>
    );
}
