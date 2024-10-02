'use strict';

function writeThreadEvents(targetThreadDivID) {
    const threadDiv = $(targetThreadDivID).first()
    const threadID = threadDiv.data("otto-thread-id");
    const es = new EventSource("/ui/threads/" + threadID  + "/events");

    // Crazy state vars
    let lastRunID = "";
    let lastRunDiv = null;
    let lastStepID = "";
    let lastStepDiv = null;

    // We do this because EventSource might resume itself in the middle of a run due to unforeseen circumstances.
    es.addEventListener("start", function (e) {
        // Reset the lastRunID so that we can clear the previous run that might have
        // been partial
        lastRunID = ""
        lastStepID = ""
    })

    // We do this to be done, because we are done, and we should stop doing things.
    es.addEventListener("close", function (e) {
        es.close();
    })

    es.onmessage = (e) => {
        const data = JSON.parse(e.data);
        const runID = data.runID || "";

        // If the runID is empty, we ignore the message
        if (runID === "") {
            return
        }

        // First run clear up old state
        if (lastRunID === "") {
            let found = false;
            for (let child of threadDiv.children()) {
                if (found) {
                    child.remove();
                } else if (child.id === runID) {
                    found = true;
                    child.remove();
                    break;
                }
            }
        }

        // Add the step container if this is a new step
        if ("step" in data) {
            if (lastStepID !== data.step.id) {
                lastStepID = data.step.id;
                lastStepDiv = threadDiv.append($("<div>", {
                    id: "step_" + data.step.id,
                    class: "step",
                })).children().last()
            }
        }

        // Add the run container if this is a new run
        if (lastRunID !== runID) {
            lastRunID = runID;
            // If we have a step container, add the run to that, otherwise add it to the thread
            lastRunDiv = (lastStepDiv || threadDiv).append($("<div>", {
                id: "run_" + runID,
                class: "run",
            })).children().last();
        }

        if (data.waitingOnModel) {
            lastRunDiv.append(messageWaiting(data.time))
            return
        }

        lastRunDiv.children(".message-waiting-on-model").remove();

        // Process mutually exclusive messages

        if (data.input) {
            lastRunDiv.append(messageInputDiv(data.input, data.time))
        } else if (data.stepTemplateInvoke) {
            const newContent = JSON.stringify(data.stepTemplateInvoke)
            const newDiv = messageInputDiv(newContent, data.time)
            newDiv.data("otto-content", newContent)
            appendToDiv(lastRunDiv, "message-input", newDiv)
        } else if (data.error) {
            lastRunDiv.append(messageErrorDiv(data.input, data.time))
        } else if (data.toolInput) {
            appendToDiv(lastRunDiv, "message-tool-input", messageToolInputDiv(data.toolInput.content, data.time))
        } else if (data.toolCall) {
            lastRunDiv.children(".message-tool-input").remove();
            lastRunDiv.append(messageToolCallDiv(data.toolCall.name, data.toolCall.input, data.time, data.runID))
        } else if (data.workflowCall) {
            lastRunDiv.children(".message-tool-input").remove();
            lastRunDiv.append(messageWorkflowCallDiv(data.workflowCall.name, data.workflowCall.input, data.time, data.runID, threadID, data.workflowCall.workflowID, data.workflowCall.threadID))
            // we added a new button
            htmx.process(document.body)
        } else if (data.content) {
            appendToDiv(lastRunDiv, "message-content", messageContentDev(data.content, data.time))
        }
    };
}

function formatTime(time) {
    if (time === null) {
        return new Date().toLocaleString()
    }
    const date = new Date(time);
    return date.toLocaleString();
}

function messageWaiting(timestamp) {
    return $(`<div class="message-waiting-on-model flex justify-end" >
        <div class="flex items-start gap-2.5 mb-6">
            <div class="flex flex-col gap-1 w-full max-w-[320px]">
                <div class="flex items-center space-x-2 rtl:space-x-reverse">
                    <span class="text-sm font-semibold text-gray-900 dark:text-white">AI</span>
                    <span class="text-sm font-normal text-gray-500 dark:text-gray-400">${formatTime(timestamp)}</span>
                </div>
                <div
                    class="flex flex-col leading-1.5 p-4 border-gray-200 bg-gray-100 rounded-e-xl rounded-es-xl dark:bg-gray-700">
                    <div class="message-text text-sm font-normal text-gray-900 dark:text-white">
                        Waiting for AI response
                    </div>
                </div>
            </div>
        </div>
    </div>`)
}

function messageContentDev(input, timestamp) {
    const el = $(`<div class="message-content flex justify-end" >
        <div class="flex items-start gap-2.5 mb-6">
            <div class="flex flex-col gap-1 w-full max-w-[320px]">
                <div class="flex items-center space-x-2 rtl:space-x-reverse">
                    <span class="text-sm font-semibold text-gray-900 dark:text-white">AI</span>
                    <span class="text-sm font-normal text-gray-500 dark:text-gray-400">${formatTime(timestamp)}</span>
                </div>
                <div
                    class="flex flex-col leading-1.5 p-4 border-gray-200 bg-gray-100 rounded-e-xl rounded-es-xl dark:bg-gray-700">
                    <div class="message-text text-sm font-normal text-gray-900 dark:text-white">
                        ${DOMPurify.sanitize(marked.parse(input))}
                    </div>
                </div>
            </div>
        </div>
    </div>`)
    el.data("otto-content", input)
    return el
}

function messageToolCallDiv(name, input, timestamp, runID) {
    if ( input === '' || input === '{}' ) {
       input = 'No input'
    }
    return $(`<div class="message-tool-call flex justify-end" >
        <div class="flex items-start gap-2.5 mb-6">
            <div class="flex flex-col gap-1 w-full max-w-[320px]">
                <div class="flex items-center space-x-2 rtl:space-x-reverse">
                    <span class="text-sm font-semibold text-gray-900 dark:text-white">Tool Call: ${_.escape(name)}</span>
                    <span class="text-sm font-normal text-gray-500 dark:text-gray-400">${formatTime(timestamp)}</span>
                </div>
                <div
                    class="flex flex-col leading-1.5 p-4 border-gray-200 bg-gray-100 rounded-e-xl rounded-es-xl dark:bg-gray-700">
                    <div class="message-text text-sm font-normal text-gray-900 dark:text-white">
                        ${_.escape(input)}
                    </div>
                </div>
                <div class="flex justify-end" >
                    <span class="text-sm font-normal text-gray-500 dark:text-gray-400">${runID}</span>
                </div>
            </div>
        </div>
    </div>`)
}

function messageWorkflowCallDiv(name, input, timestamp, runID, currentThreadID, workflowID, nextThreadID) {
    const el = $(`<div class="message-workflow-call flex justify-end" >
        <div class="flex items-start gap-2.5 mb-6">
            <div class="flex flex-col gap-1 w-full max-w-[320px]">
                <div class="flex items-center space-x-2 rtl:space-x-reverse">
                    <span class="text-sm font-semibold text-gray-900 dark:text-white">Workflow Call: ${_.escape(name)}</span>
                    <span class="text-sm font-normal text-gray-500 dark:text-gray-400">${formatTime(timestamp)}</span>
                </div>
                <button type="button"
                    class="text-gray-900 bg-white border border-gray-300 focus:outline-none hover:bg-gray-100 focus:ring-4 focus:ring-gray-100 font-medium rounded-lg text-sm px-5 py-2.5 me-2 mb-2 dark:bg-gray-800 dark:text-white dark:border-gray-600 dark:hover:bg-gray-700 dark:hover:border-gray-600 dark:focus:ring-gray-700"
                    hx-target="#${currentThreadID}-next-thread"
                    hx-get="/ui/workflows/${workflowID}/threads/${nextThreadID}"
                    >
                    Details
                </button>
                <div class="flex justify-end" >
                    <span class="text-sm font-normal text-gray-500 dark:text-gray-400">${runID}</span>
                </div>
            </div>
        </div>
    </div>`)
    el.data("otto-content", input)
    return el
}

function messageToolInputDiv(input, timestamp) {
    const el = $(`<div class="message-tool-input flex justify-end" >
        <div class="flex items-start gap-2.5 mb-6">
            <div class="flex flex-col gap-1 w-full max-w-[320px]">
                <div class="flex items-center space-x-2 rtl:space-x-reverse">
                    <span class="text-sm font-semibold text-gray-900 dark:text-white">Generating Tool Call Input</span>
                    <span class="text-sm font-normal text-gray-500 dark:text-gray-400">${formatTime(timestamp)}</span>
                </div>
                <div
                    class="flex flex-col leading-1.5 p-4 border-gray-200 bg-gray-100 rounded-e-xl rounded-es-xl dark:bg-gray-700">
                    <div class="message-text text-sm font-normal text-gray-900 dark:text-white">
                        ${_.escape(input)}
                    </div>
                </div>
            </div>
        </div>
    </div>`)
    el.data("otto-content", input)
    return el
}

function messageInputDiv(input, timestamp) {
    return $(`<div class="message-input flex items-start gap-2.5 mb-6">
        <div class="flex flex-col gap-1 w-full max-w-[320px]">
            <div class="flex items-center space-x-2 rtl:space-x-reverse">
                <span class="text-sm font-semibold text-gray-900 dark:text-white">Step Input</span>
                <span class="text-sm font-normal text-gray-500 dark:text-gray-400">${formatTime(timestamp)}</span>
            </div>
            <div
                class="flex flex-col leading-1.5 p-4 border-gray-200 bg-gray-100 rounded-e-xl rounded-es-xl dark:bg-gray-700">
                <div class="message-text text-sm font-normal text-gray-900 dark:text-white">
                    ${_.escape(input)}
                </div>
            </div>
        </div>
    </div>`)
}

function messageErrorDiv(input, timestamp) {
    return $(`<div class="message-error flex justify-end" >
        <div class="flex items-start gap-2.5 mb-6">
            <div class="flex flex-col gap-1 w-full max-w-[320px]">
                <div class="flex items-center space-x-2 rtl:space-x-reverse">
                    <span class="text-sm font-semibold text-gray-900 dark:text-white">ERROR</span>
                    <span class="text-sm font-normal text-gray-500 dark:text-gray-400">${formatTime(timestamp)}</span>
                </div>
                <div
                    class="flex flex-col leading-1.5 p-4 border-gray-200 bg-gray-100 rounded-e-xl rounded-es-xl dark:bg-gray-700">
                    <div class="message-text text-sm font-normal text-gray-900 dark:text-white">
                        ${_.escape(input)}
                    </div>
                </div>
            </div>
        </div>
    </div>`)
}

function appendToDiv(lastDiv, divClass, newDiv) {
    const last = lastDiv.children().last()
    if (last.hasClass(divClass)) {
        const newContent = (last.data("otto-content") || "") + newDiv.data("otto-content")
        last.data("otto-content", newContent)
        const t = DOMPurify.sanitize(marked.parse(newContent))
        last.find('.message-text').html(t)
    } else {
        // Otherwise we append a new div
        lastDiv.append(newDiv)
    }
}