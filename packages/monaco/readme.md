# Embeddable Monaco

> Monaco Editor instance running in an embeddable iframe with an easy-to-use messaging and parameter API

Test on [https://embeddable-monaco.lukasbach.com](https://embeddable-monaco.lukasbach.com).

[Monaco Editor](https://microsoft.github.io/monaco-editor/) is an code editor that powers VS Code and can be reused
in other apps. Compared to other frontend libraries, setup is a bit harder with Monaco since it uses webworkers and
requires a more complicated build setup. In most use cases it can be implemented with build plugins, but in some
use cases, there are harder restrictions on the availability of the build system or the use of web workers, where
this is not possible.

This project provides a pre-built version of Monaco Editor that can be embedded in an iframe and controlled via
a simple messaging API. It is a drop-in replacement for the Monaco Editor, but it can be used in environments
where the build system is not available or where web workers are not allowed.

## Example

```html
<iframe src="https://embeddable-monaco.lukasbach.com?code=initial_code&language=javascript&theme=vs-dark" id="iframe"></iframe>
```

```javascript
const iframe = document.getElementById("iframe");
window.addEventListener("message", ({ data }) => {
    switch(data.type) {
        case "ready": {
            iframe.contentWindow?.postMessage({
                type: "change-value",
                value: "new initial value"
            }, "*");
            break;
        }
        case "change": {
            console.log("current value is", data.value)
            break;
        }
    }
});
```


## Usage

To use, just embed the URL https://embeddable-monaco.lukasbach.com in an iframe.

There are two primary ways to configure and communicate with the editor. First, configuration can be passed through
URL query parameters. Second, the editor can be controlled via a simple messaging API, so that configuration
can be changed during runtime, and changes to the editor can be passed back to the parent frame.

The implementation is not complicated, so an easy way to understand what is possible is to just look into
the implementation file [src/embed.ts](src/embed.ts).

## Supported query parameters

The following query parameters are supported:

- `code`: Initial code, defaults to empty
- `lang`: Initial language, defaults to javascript
- `theme`: Initial theme, defaults to vs-light
  - Also supports all of these themes: https://github.com/brijeshb42/monaco-themes/blob/master/themes/themelist.json
  - Use the value of the theme as name, e.g. "Vibrant Ink"
- `contextmenu`: boolean, "true" or "false"
- `folding`: boolean, "true" or "false"
- `readonly`: boolean, "true" or "false"
- `lineNumbers`: boolean, "on" or "off"
- `minimap`: boolean, "true" or "false"
- `background`: custom background color, can also be transparent
- `javascriptDefaults`: set javascript language properties, required to use:
  - `javascriptDefaultsNoSemanticValidation`
  - `javascriptDefaultsNoSyntaxValidation`
- `typescriptDefaults`: set typescript language properties, required to use:
  - `typescriptDefaultsNoSemanticValidation`
  - `typescriptDefaultsNoSyntaxValidation`
- `dontPostValueOnChange`: In the `change` handler, don't post the value back to the parent frame every time the model is changed
- `context`: a string that is passed back in every message sent back from the iframe

## Messaging API

### Messages sent by the iframe to the parent

Receive messages via

```javascript
iframe.addEventListener('message', (e) => {
    console.log(`message type is ${e.data.type}, payload is in`, e.data);
});
```

Messages sent by the iframe:

- `{ type: "ready" }`: Sent when the editor is ready
- `{ type: "change", value: string }`: Sent everytime the editor value changes
- `{ type: "content", value: string }`: Sent when the editor content is requested via `get-content` message

### Messages that can be sent to the iframe from the parent

Send messages to the iframe via

```javascript
iframe.contentWindow.postMessage({ type: "type", ...parameters }, "*");
```

Messages that can be sent to the iframe

- `{ type: "change-options", options: IEditorOptions }`: Change editor options
- `{ type: "change-value", value: string }`: Change editor value
- `{ type: "change-language", language: string }`: Change editor language
- `{ type: "change-theme", theme: string }`: Change editor theme
- `{ type: "change-background", background: string, theme?: string }`: Change editor background color. Can also be transparent
- `{ type: "change-javascript-defaults", javascriptDefaults: IJavaScriptDefaults }`: Change javascript language defaults
- `{ type: "change-typescript-defaults", typescriptDefaults: ITypeScriptDefaults }`: Change typescript language defaults
- `{ type: "get-content }`: IFrame will dispatch a `content` message with the current editor value
