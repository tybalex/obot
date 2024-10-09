'use strict';

let chat;

document.addEventListener('alpine:init', () => {
    chat = new Chat();
    Alpine.store('chat', {
        messages: [],
        files: [],
    });
})

document.addEventListener('alpine:initialized', () => {
    chat.start();
})

class Chat {
    constructor() {
        this.running = false
    }

    start() {
        if (this.running) {
            return;
        }
        this.files = Alpine.store('chat').files
        this.messages = Alpine.store('chat').messages
        this.es = new EventSource('/events');
        this.es.onmessage = this.onMessage.bind(this);
        this.es.onerror = this.onError.bind(this);
        this.running = true
    }

    stop() {
        if (this.es) {
            this.es.close();
            this.es = null;
        }
        this.running = false
        setTimeout(() => this.start(), 5000)
    }

    submit(message) {
        message.changed = []
        const onSuccess = []
        for (const file of Alpine.store('chat').files) {
            if (file.partial) {
                continue
            }

            if (file.content !== file.original) {
                message.changed.push({
                    filename: file.name,
                    content: file.content,
                })
                onSuccess.push(() => {
                    file.original = file.content
                })
            }


        }
        htmx.ajax('POST', '/chat', {
            values: {
                message: message,
            },
            target: '#status',
        }).then(() => {
            for (const callback of onSuccess) {
                callback()
            }
        })
    }

    onError(e) {
        console.error(e);
        this.stop();
    }

    appendMessage(msg) {
        if (this.messages.length > 0) {
            this.messages.at(-1).done = true;
        }
        this.messages.push(msg);
    }

    pushContent(event) {
        const content = event.toolInput ? event.toolInput.input : event.content;
        if (!content) {
            return;
        }
        const last = this.messages.at(-1);
        if (last && last.contentID === event.contentID) {
            last.content.push(content);
            const fileWrite = this.parseFileWrite(event)
            if (fileWrite) {
                this.setFileContent(fileWrite.filename, last.content.join(''), true)
            }
        } else {
            this.appendMessage({
                ...event,
                content: [content],
            });
        }
    }

    setFileContent(name, content, partial = false) {
        const file = this.files.find(f => f.name === name)
        if (file) {
            file.partial = partial
            file.content = content
            if (!partial) {
                file.original = content
            }
        } else {
            this.files.push({
                name: name,
                original: content,
                content: content,
                partial: partial,
            })
            if (!partial) {
                htmx.trigger("#chat-sidebar", "files-changed", {})
            }
        }
    }

    onMessage(msg) {
        const event = JSON.parse(msg.data);
        if (!"runID" in event) {
            return;
        }

        if (event.input) {
            // Clear the input. This is a bit of a hack.
            event.content = ""
        }

        if (event.contentID) {
            this.pushContent(event);
        } else {
            this.appendMessage(event);
        }

        document.getElementById('messages-scroll').dispatchEvent(new CustomEvent('message-added'));

        const fileWrite = this.parseFileWrite(event)
        if (fileWrite && !fileWrite.partial) {
            this.setFileContent(fileWrite.filename, fileWrite.content)
        }
    }

    parseFileWrite(event) {
        const target = event.toolCall || event.toolInput
        const partial = event.toolCall === undefined
        if (target && target.name === "workspace_write") {
            try {
                let parsed = JSON.parse(target.input)
                return {
                    partial: partial,
                    ...parsed,
                }
            } catch (e) {
                try {
                    let parsed = JSON.parse(target.input + '"}')
                    return {
                        partial: true,
                        ...parsed,
                    }
                } catch (e) {
                }
            }
        }
    }

    copySelection(el) {
        const selection = window.getSelection()
        if (selection.anchorNode && 'filename' in selection.anchorNode.dataset) {
            el.dataset.filename = selection.anchorNode.dataset.filename
            el.dataset.selection = selection.toString()
        }
    }

    explain(el) {
        const filename = el.dataset.filename
        const selection = el.dataset.selection
        if (selection.length === 0) {
            return
        }
        this.submit({explain: {filename, selection}})
    }

    improve(el) {
        const filename = el.dataset.filename
        const selection = el.dataset.selection
        if (selection.length === 0) {
            return
        }
        this.submit({
            prompt: el.value,
            improve: {filename, selection}
        })
    }
}

function getFilenameAndSelection(el) {
    const selection = window.getSelection()
    if (selection.anchorNode && 'filename' in selection.anchorNode.dataset) {
        return {
            filename: selection.anchorNode.dataset.filename,
            selection: selection.toString()
        }
    }
    return null
}

function explainSelection(el) {
    const selection = getFilenameAndSelection(el)
    if (selection) {
        chat.submit({explain: selection})
    }
}


function scrollThis(el) {
    el.scrollTo({
        top: el.scrollHeight,
        behavior: 'smooth'
    })
}