import { Suspense } from "react";
import { EditorPageClient } from "@/components/editor/editor-page-client";

export default function EditorPage() {
  return (
    <Suspense
      fallback={<main className="editor-loading">Loading editor...</main>}
    >
      <EditorPageClient />
    </Suspense>
  );
}
