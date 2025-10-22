import { defineConfig } from "vite";
import monacoEditorPlugin from "vite-plugin-monaco-editor";

console.log(monacoEditorPlugin);

// https://vitejs.dev/config/
export default defineConfig({
  base: "/monaco/",
  plugins: [(monacoEditorPlugin as any).default({})],
  build: {
    outDir: "../../frontend/public/monaco",
    emptyOutDir: false,
  },
});
