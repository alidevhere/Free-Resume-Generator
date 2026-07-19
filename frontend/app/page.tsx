"use client";

import { useEffect, useRef, useState } from "react";
import { useRouter } from "next/navigation";
import {
  createResume,
  deleteResume,
  listResumes,
  saveResume,
} from "@/lib/api-client";
import {
  createEmptyResume,
  DEFAULT_SECTION_ORDER,
} from "@/lib/resume-defaults";
import { Resume, ResumeListItem } from "@/lib/types";
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
  const fileInputRef = useRef<HTMLInputElement | null>(null);

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

  function handleImportJson(file: File) {
    const reader = new FileReader();
    reader.onload = async () => {
      try {
        const parsed = JSON.parse(String(reader.result)) as Partial<
          Resume & { resume?: Resume; name?: string }
        >;

        // Support both a bare Resume object and the { name, resume } envelope.
        const candidate: Resume | undefined =
          parsed.resume ?? (parsed as Resume);

        if (!candidate || typeof candidate !== "object") {
          throw new Error("Invalid resume JSON: expected an object.");
        }

        if (typeof candidate.name !== "string") {
          throw new Error("Invalid resume JSON: missing string field 'name'.");
        }

        const normalized: Resume = {
          ...createEmptyResume(),
          ...candidate,
          sectionOrder: candidate.sectionOrder?.length
            ? candidate.sectionOrder
            : [...DEFAULT_SECTION_ORDER],
          experience: candidate.experience || [],
          education: candidate.education || [],
          projects: candidate.projects || [],
          skills: candidate.skills || [],
          certifications: candidate.certifications || [],
          openSourceContributions: candidate.openSourceContributions || [],
        };

        const recordName =
          (parsed.name && typeof parsed.name === "string" && parsed.name) ||
          candidate.name ||
          "Imported Resume";

        const record = await saveResume(null, {
          name: recordName,
          resume: normalized,
        });

        setToast({
          tone: "success",
          message: "Resume imported successfully.",
        });
        router.push(`/editor?id=${record.id}`);
      } catch (error) {
        setToast({
          tone: "error",
          message:
            error instanceof Error
              ? error.message
              : "Failed to import resume JSON",
        });
      }
    };
    reader.onerror = () => {
      setToast({
        tone: "error",
        message: "Failed to read the selected file.",
      });
    };
    reader.readAsText(file);
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
            <div className="library-action-buttons">
              <input
                ref={fileInputRef}
                type="file"
                accept="application/json,.json"
                style={{ display: "none" }}
                onChange={(e) => {
                  const file = e.target.files?.[0];
                  if (file) {
                    handleImportJson(file);
                  }
                  // Reset so selecting the same file again re-triggers change.
                  e.target.value = "";
                }}
              />
              <Button
                variant="secondary"
                onClick={() => fileInputRef.current?.click()}
              >
                Import JSON
              </Button>
              <Button onClick={handleCreate}>Create New Resume</Button>
            </div>
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
