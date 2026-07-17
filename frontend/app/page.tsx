"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { createResume, deleteResume, listResumes } from "@/lib/api-client";
import { ResumeListItem } from "@/lib/types";
import { ResumeCard } from "@/components/library/resume-card";
import { Button } from "@/components/ui/button";
import { Toast } from "@/components/ui/toast";

type ToastState = {
  tone: "success" | "error";
  message: string;
} | null;

export default function LibraryPage() {
  const router = useRouter();
  const [items, setItems] = useState<ResumeListItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [toast, setToast] = useState<ToastState>(null);

  useEffect(() => {
    void loadItems();
  }, []);

  useEffect(() => {
    if (!toast) {
      return;
    }
    const timer = window.setTimeout(() => setToast(null), 3000);
    return () => window.clearTimeout(timer);
  }, [toast]);

  async function loadItems() {
    try {
      setLoading(true);
      const response = await listResumes();
      setItems(response);
    } catch (error) {
      setToast({
        tone: "error",
        message:
          error instanceof Error ? error.message : "Failed to load resumes",
      });
    } finally {
      setLoading(false);
    }
  }

  async function handleCreate() {
    try {
      const record = await createResume();
      router.push(`/editor?id=${record.id}`);
    } catch (error) {
      setToast({
        tone: "error",
        message:
          error instanceof Error ? error.message : "Failed to create resume",
      });
    }
  }

  async function handleDelete(id: number, name: string) {
    const shouldDelete = window.confirm(
      `Delete resume \"${name}\"? This action cannot be undone.`,
    );
    if (!shouldDelete) {
      return;
    }

    try {
      await deleteResume(id);
      setToast({ tone: "success", message: "Resume deleted successfully." });
      await loadItems();
    } catch (error) {
      setToast({
        tone: "error",
        message:
          error instanceof Error ? error.message : "Failed to delete resume",
      });
    }
  }

  return (
    <main className="library-shell">
      <div className="library-panel">
        <header className="library-header">
          <h1>Free Resume Generator</h1>
          <p>Select an existing resume or create a new one to get started.</p>
        </header>
        <div className="library-content">
          <div className="library-actions">
            <strong>Your Resumes</strong>
            <Button onClick={handleCreate}>Create New Resume</Button>
          </div>
          {loading ? (
            <div className="empty-library">Loading resumes...</div>
          ) : null}
          {!loading && items.length === 0 ? (
            <div className="empty-library">
              <h3>No resumes yet</h3>
              <p>
                Create your first resume to start editing and generating PDFs.
              </p>
              <Button onClick={handleCreate}>Create New Resume</Button>
            </div>
          ) : null}
          {!loading && items.length > 0 ? (
            <div className="resume-grid">
              {items.map((item) => (
                <ResumeCard
                  key={item.id}
                  item={item}
                  onOpen={(id) => router.push(`/editor?id=${id}`)}
                  onDelete={handleDelete}
                />
              ))}
            </div>
          ) : null}
        </div>
      </div>
      {toast ? <Toast tone={toast.tone} message={toast.message} /> : null}
    </main>
  );
}
