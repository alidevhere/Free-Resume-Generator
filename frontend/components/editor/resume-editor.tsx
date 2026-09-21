"use client";

import { useCallback, useEffect, useMemo, useState } from "react";
import { useRouter } from "next/navigation";
import {
  buildLatex,
  buildPdf,
  createResume,
  getResume,
  saveResume,
} from "@/lib/api-client";
import {
  createEmptyResume,
  DEFAULT_SECTION_ORDER,
} from "@/lib/resume-defaults";
import { joinCsv, joinMultiline, splitCsv, splitMultiline } from "@/lib/format";
import {
  Certification,
  Education,
  Experience,
  OpenSourceContribution,
  Project,
  Resume,
  SkillCategory,
} from "@/lib/types";
import { Button } from "@/components/ui/button";
import { ItemListEditor } from "@/components/editor/item-list-editor";
import { SectionOrderEditor } from "@/components/editor/section-order-editor";
import { SectionShell } from "@/components/editor/section-shell";
import { Toast } from "@/components/ui/toast";

type Props = {
  resumeId: number | null;
};

type ToastState = {
  tone: "success" | "error";
  message: string;
} | null;

type DownloadFormat = "pdf" | "latex" | "json";

export function ResumeEditor({ resumeId }: Props) {
  const router = useRouter();
  const [recordId, setRecordId] = useState<number | null>(resumeId);
  const [recordName, setRecordName] = useState("Untitled Resume");
  const [createdAt, setCreatedAt] = useState("");
  const [updatedAt, setUpdatedAt] = useState("");
  const [resume, setResume] = useState<Resume>(createEmptyResume());
  const [loading, setLoading] = useState(true);
  const [building, setBuilding] = useState(false);
  const [statusText, setStatusText] = useState("Waiting for first build");
  const [toast, setToast] = useState<ToastState>(null);
  const [previewUrl, setPreviewUrl] = useState("");
  const [pdfBlob, setPdfBlob] = useState<Blob | null>(null);
  const [downloadFormat, setDownloadFormat] = useState<DownloadFormat>("pdf");

  const initialize = useCallback(async () => {
    try {
      setLoading(true);
      if (resumeId === null) {
        const record = await createResume();
        hydrateFromRecord(record);
        return;
      }

      const record = await getResume(resumeId);
      hydrateFromRecord(record);
    } catch (error) {
      setToast({
        tone: "error",
        message:
          error instanceof Error ? error.message : "Failed to load resume",
      });
    } finally {
      setLoading(false);
    }
  }, [resumeId]);

  useEffect(() => {
    void initialize();
  }, [initialize]);

  useEffect(() => {
    if (!toast) {
      return;
    }
    const timer = window.setTimeout(() => setToast(null), 3000);
    return () => window.clearTimeout(timer);
  }, [toast]);

  useEffect(() => {
    return () => {
      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
      }
    };
  }, [previewUrl]);

  function hydrateFromRecord(record: {
    id: number;
    name: string;
    createdAt: string;
    updatedAt: string;
    resume: Resume;
  }) {
    setRecordId(record.id);
    setRecordName(record.name || "Untitled Resume");
    setCreatedAt(record.createdAt || "");
    setUpdatedAt(record.updatedAt || "");

    const normalized = {
      ...createEmptyResume(),
      ...record.resume,
      experience: record.resume.experience || [],
      education: record.resume.education || [],
      projects: record.resume.projects || [],
      skills: record.resume.skills || [],
      certifications: record.resume.certifications || [],
      openSourceContributions: record.resume.openSourceContributions || [],
      sectionOrder: record.resume.sectionOrder?.length
        ? record.resume.sectionOrder
        : [...DEFAULT_SECTION_ORDER],
    };
    setResume(normalized);
  }

  function patchResume(next: Partial<Resume>) {
    setResume((current) => ({ ...current, ...next }));
  }

  function patchList<T>(
    key:
      | "experience"
      | "education"
      | "projects"
      | "skills"
      | "certifications"
      | "openSourceContributions",
    updater: (items: T[]) => T[],
  ) {
    setResume((current) => ({
      ...current,
      [key]: updater(current[key] as T[]),
    }));
  }

  async function persist(silent = false) {
    const payload = {
      name: recordName.trim() || "Untitled Resume",
      resume,
    };

    const record = await saveResume(recordId, payload);
    setRecordId(record.id);
    setCreatedAt(record.createdAt);
    setUpdatedAt(record.updatedAt);
    if (!silent) {
      setToast({ tone: "success", message: "Resume saved successfully." });
    }
  }

  function downloadBlob(blob: Blob, filename: string) {
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = filename;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(url);
  }

  async function handleBuildAndPreview() {
    if (building) {
      return;
    }

    if (!resume.name.trim() || !resume.email.trim()) {
      setToast({ tone: "error", message: "Name and email are required." });
      return;
    }

    try {
      setBuilding(true);
      setStatusText("Building PDF...");
      await persist(true);
      const blob = await buildPdf(resume);
      setPdfBlob(blob);
      setStatusText(`Preview updated at ${new Date().toLocaleTimeString()}`);

      if (previewUrl) {
        URL.revokeObjectURL(previewUrl);
      }
      const nextUrl = URL.createObjectURL(blob);
      setPreviewUrl(nextUrl);
      setToast({ tone: "success", message: "Preview updated successfully." });
    } catch (error) {
      setStatusText("Build failed");
      setToast({
        tone: "error",
        message: error instanceof Error ? error.message : "Failed to build PDF",
      });
    } finally {
      setBuilding(false);
    }
  }

  async function handleDownloadPdf() {
    try {
      if (pdfBlob) {
        downloadBlob(pdfBlob, "resume.pdf");
        setToast({ tone: "success", message: "PDF downloaded successfully." });
        return;
      }

      await handleBuildAndPreview();
      if (pdfBlob) {
        downloadBlob(pdfBlob, "resume.pdf");
      }
    } catch (error) {
      setToast({
        tone: "error",
        message:
          error instanceof Error ? error.message : "Failed to download PDF",
      });
    }
  }

  async function handleDownloadLatex() {
    if (building) {
      return;
    }

    try {
      setBuilding(true);
      setStatusText("Generating LaTeX...");
      const blob = await buildLatex(resume);
      downloadBlob(blob, "resume.tex");
      setStatusText("LaTeX generated");
      setToast({ tone: "success", message: "LaTeX downloaded successfully." });
    } catch (error) {
      setStatusText("LaTeX generation failed");
      setToast({
        tone: "error",
        message:
          error instanceof Error ? error.message : "Failed to generate LaTeX",
      });
    } finally {
      setBuilding(false);
    }
  }

  function handleDownloadJson() {
    const payload = {
      name: recordName.trim() || "Untitled Resume",
      resume,
    };
    const blob = new Blob([JSON.stringify(payload, null, 2)], {
      type: "application/json",
    });
    downloadBlob(blob, "resume.json");
    setToast({ tone: "success", message: "JSON downloaded successfully." });
  }

  async function handleDownload() {
    if (downloadFormat === "pdf") {
      await handleDownloadPdf();
    } else if (downloadFormat === "latex") {
      await handleDownloadLatex();
    } else {
      handleDownloadJson();
    }
  }

  const previewHref = useMemo(() => previewUrl || "#", [previewUrl]);

  if (loading) {
    return <main className="editor-loading">Loading editor...</main>;
  }

  return (
    <main className="editor-container">
      <header className="header">
        <h1>Free Resume Generator</h1>
        <p>Build and preview your resume side by side.</p>
      </header>

      <div className="workspace">
        <section className="preview-pane">
          <div className="preview-toolbar">
            <span>PDF Preview</span>
            <span className="preview-status">{statusText}</span>
            <a
              href={previewHref}
              className="preview-link"
              target="_blank"
              rel="noreferrer noopener"
            >
              Open in new tab
            </a>
          </div>
          <div className="preview-frame-wrap">
            {previewUrl ? (
              <object
                className="pdf-preview"
                type="application/pdf"
                data={previewUrl}
              >
                <div className="preview-fallback">
                  Preview not available in this browser.
                </div>
              </object>
            ) : (
              <div className="preview-fallback">
                Build your resume to generate a preview.
              </div>
            )}
          </div>
        </section>

        <aside className="form-pane">
          <div className="form-pane-inner">
            <div className="build-bar">
              <Button variant="slate" onClick={() => router.push("/")}>
                Back
              </Button>
              <Button
                variant="primary"
                onClick={() => void persist()}
                disabled={building}
              >
                Save Resume
              </Button>
              <Button
                variant="primary"
                onClick={() => void handleBuildAndPreview()}
                disabled={building}
              >
                Build and Preview
              </Button>
              <div className="download-group">
                <select
                  id="downloadFormat"
                  value={downloadFormat}
                  onChange={(e) =>
                    setDownloadFormat(e.target.value as DownloadFormat)
                  }
                  aria-label="Download format"
                >
                  <option value="pdf">PDF</option>
                  <option value="latex">LaTeX</option>
                  <option value="json">JSON</option>
                </select>
                <Button
                  variant="success"
                  onClick={() => void handleDownload()}
                  disabled={building}
                >
                  Download
                </Button>
              </div>
            </div>

            <div className="resume-meta-bar">
              <div className="form-group">
                <label htmlFor="resumeName">Resume Name</label>
                <input
                  id="resumeName"
                  value={recordName}
                  onChange={(event) => setRecordName(event.target.value)}
                />
              </div>
              <div className="meta-text-stack">
                <div className="meta-text">Created: {createdAt || "-"}</div>
                <div className="meta-text">Updated: {updatedAt || "-"}</div>
              </div>
            </div>

            <SectionShell title="Resume Template">
              <div className="form-group">
                <label htmlFor="template">Template</label>
                <select
                  id="template"
                  value={resume.template}
                  onChange={(event) =>
                    patchResume({ template: event.target.value })
                  }
                >
                  <option value="enhanced-faang-resume">
                    Enhanced FAANG Resume
                  </option>
                </select>
              </div>
            </SectionShell>

            <SectionShell title="Section Order">
              <SectionOrderEditor
                order={resume.sectionOrder}
                onChange={(order) => patchResume({ sectionOrder: order })}
              />
            </SectionShell>

            <SectionShell title="Personal Information">
              <div className="form-group">
                <label htmlFor="name">Full Name</label>
                <input
                  id="name"
                  value={resume.name}
                  onChange={(e) => patchResume({ name: e.target.value })}
                />
              </div>
              <div className="two-column">
                <div className="form-group">
                  <label htmlFor="email">Email</label>
                  <input
                    id="email"
                    type="email"
                    value={resume.email}
                    onChange={(e) => patchResume({ email: e.target.value })}
                  />
                </div>
                <div className="form-group">
                  <label htmlFor="phone">Phone</label>
                  <input
                    id="phone"
                    value={resume.phone}
                    onChange={(e) => patchResume({ phone: e.target.value })}
                  />
                </div>
              </div>
              <div className="two-column">
                <div className="form-group">
                  <label htmlFor="website">Website</label>
                  <input
                    id="website"
                    value={resume.website}
                    onChange={(e) => patchResume({ website: e.target.value })}
                  />
                </div>
                <div className="form-group">
                  <label htmlFor="github">GitHub</label>
                  <input
                    id="github"
                    value={resume.github}
                    onChange={(e) => patchResume({ github: e.target.value })}
                  />
                </div>
              </div>
              <div className="form-group">
                <label htmlFor="linkedin">LinkedIn</label>
                <input
                  id="linkedin"
                  value={resume.linkedin}
                  onChange={(e) => patchResume({ linkedin: e.target.value })}
                />
              </div>
            </SectionShell>

            <SectionShell title="Summary">
              <div className="form-group">
                <label htmlFor="summary">Professional Summary</label>
                <textarea
                  id="summary"
                  value={resume.summary}
                  onChange={(e) => patchResume({ summary: e.target.value })}
                />
              </div>
            </SectionShell>

            <SectionShell title="Experience">
              <ItemListEditor<Experience>
                items={resume.experience}
                addLabel="Add Experience"
                onAdd={() =>
                  patchList<Experience>("experience", (items) => [
                    ...items,
                    {
                      company: "",
                      title: "",
                      location: "",
                      start: "",
                      end: "",
                      bullets: [],
                    },
                  ])
                }
                onRemove={(index) =>
                  patchList<Experience>("experience", (items) =>
                    items.filter((_, i) => i !== index),
                  )
                }
                onChangeItem={(index, value) =>
                  patchList<Experience>("experience", (items) =>
                    items.map((item, i) => (i === index ? value : item)),
                  )
                }
                renderItem={(item, _index, onChange) => (
                  <>
                    <div className="form-group">
                      <label>Company</label>
                      <input
                        value={item.company}
                        onChange={(e) =>
                          onChange({ ...item, company: e.target.value })
                        }
                      />
                    </div>
                    <div className="two-column">
                      <div className="form-group">
                        <label>Title</label>
                        <input
                          value={item.title}
                          onChange={(e) =>
                            onChange({ ...item, title: e.target.value })
                          }
                        />
                      </div>
                      <div className="form-group">
                        <label>Location</label>
                        <input
                          value={item.location}
                          onChange={(e) =>
                            onChange({ ...item, location: e.target.value })
                          }
                        />
                      </div>
                    </div>
                    <div className="two-column">
                      <div className="form-group">
                        <label>Start</label>
                        <input
                          value={item.start}
                          onChange={(e) =>
                            onChange({ ...item, start: e.target.value })
                          }
                        />
                      </div>
                      <div className="form-group">
                        <label>End</label>
                        <input
                          value={item.end}
                          onChange={(e) =>
                            onChange({ ...item, end: e.target.value })
                          }
                        />
                      </div>
                    </div>
                    <div className="form-group">
                      <label>Bullets (one per line)</label>
                      <textarea
                        value={joinMultiline(item.bullets)}
                        onChange={(e) =>
                          onChange({
                            ...item,
                            bullets: splitMultiline(e.target.value),
                          })
                        }
                      />
                    </div>
                  </>
                )}
              />
            </SectionShell>

            <SectionShell title="Education">
              <ItemListEditor<Education>
                items={resume.education}
                addLabel="Add Education"
                onAdd={() =>
                  patchList<Education>("education", (items) => [
                    ...items,
                    {
                      school: "",
                      degree: "",
                      location: "",
                      start: "",
                      end: "",
                    },
                  ])
                }
                onRemove={(index) =>
                  patchList<Education>("education", (items) =>
                    items.filter((_, i) => i !== index),
                  )
                }
                onChangeItem={(index, value) =>
                  patchList<Education>("education", (items) =>
                    items.map((item, i) => (i === index ? value : item)),
                  )
                }
                renderItem={(item, _index, onChange) => (
                  <>
                    <div className="form-group">
                      <label>School</label>
                      <input
                        value={item.school}
                        onChange={(e) =>
                          onChange({ ...item, school: e.target.value })
                        }
                      />
                    </div>
                    <div className="two-column">
                      <div className="form-group">
                        <label>Degree</label>
                        <input
                          value={item.degree}
                          onChange={(e) =>
                            onChange({ ...item, degree: e.target.value })
                          }
                        />
                      </div>
                      <div className="form-group">
                        <label>Location</label>
                        <input
                          value={item.location}
                          onChange={(e) =>
                            onChange({ ...item, location: e.target.value })
                          }
                        />
                      </div>
                    </div>
                    <div className="two-column">
                      <div className="form-group">
                        <label>Start</label>
                        <input
                          value={item.start}
                          onChange={(e) =>
                            onChange({ ...item, start: e.target.value })
                          }
                        />
                      </div>
                      <div className="form-group">
                        <label>End</label>
                        <input
                          value={item.end}
                          onChange={(e) =>
                            onChange({ ...item, end: e.target.value })
                          }
                        />
                      </div>
                    </div>
                  </>
                )}
              />
            </SectionShell>

            <SectionShell title="Open Source Contributions">
              <ItemListEditor<OpenSourceContribution>
                items={resume.openSourceContributions}
                addLabel="Add Contribution"
                onAdd={() =>
                  patchList<OpenSourceContribution>(
                    "openSourceContributions",
                    (items) => [
                      ...items,
                      {
                        name: "",
                        description: "",
                        role: "",
                        contribution: "",
                        link: "",
                        repoLink: "",
                        stars: "",
                        keywords: [],
                      },
                    ],
                  )
                }
                onRemove={(index) =>
                  patchList<OpenSourceContribution>(
                    "openSourceContributions",
                    (items) => items.filter((_, i) => i !== index),
                  )
                }
                onChangeItem={(index, value) =>
                  patchList<OpenSourceContribution>(
                    "openSourceContributions",
                    (items) =>
                      items.map((item, i) => (i === index ? value : item)),
                  )
                }
                renderItem={(item, _index, onChange) => (
                  <>
                    <div className="form-group">
                      <label>Project Name</label>
                      <input
                        value={item.name}
                        onChange={(e) =>
                          onChange({ ...item, name: e.target.value })
                        }
                      />
                    </div>
                    <div className="form-group">
                      <label>Description</label>
                      <input
                        value={item.description}
                        onChange={(e) =>
                          onChange({ ...item, description: e.target.value })
                        }
                      />
                    </div>
                    <div className="two-column">
                      <div className="form-group">
                        <label>Role</label>
                        <input
                          value={item.role}
                          onChange={(e) =>
                            onChange({ ...item, role: e.target.value })
                          }
                        />
                      </div>
                      <div className="form-group">
                        <label>Stars</label>
                        <input
                          value={item.stars}
                          onChange={(e) =>
                            onChange({ ...item, stars: e.target.value })
                          }
                        />
                      </div>
                    </div>
                    <div className="form-group">
                      <label>Contribution</label>
                      <textarea
                        value={item.contribution}
                        onChange={(e) =>
                          onChange({ ...item, contribution: e.target.value })
                        }
                      />
                    </div>
                    <div className="two-column">
                      <div className="form-group">
                        <label>Project Link</label>
                        <input
                          value={item.link}
                          onChange={(e) =>
                            onChange({ ...item, link: e.target.value })
                          }
                        />
                      </div>
                      <div className="form-group">
                        <label>Repo Link</label>
                        <input
                          value={item.repoLink}
                          onChange={(e) =>
                            onChange({ ...item, repoLink: e.target.value })
                          }
                        />
                      </div>
                    </div>
                    <div className="form-group">
                      <label>Keywords (comma-separated)</label>
                      <input
                        value={joinCsv(item.keywords)}
                        onChange={(e) =>
                          onChange({
                            ...item,
                            keywords: splitCsv(e.target.value),
                          })
                        }
                      />
                    </div>
                  </>
                )}
              />
            </SectionShell>

            <SectionShell title="Projects">
              <ItemListEditor<Project>
                items={resume.projects}
                addLabel="Add Project"
                onAdd={() =>
                  patchList<Project>("projects", (items) => [
                    ...items,
                    { name: "", link: "", technologies: [], bullets: [] },
                  ])
                }
                onRemove={(index) =>
                  patchList<Project>("projects", (items) =>
                    items.filter((_, i) => i !== index),
                  )
                }
                onChangeItem={(index, value) =>
                  patchList<Project>("projects", (items) =>
                    items.map((item, i) => (i === index ? value : item)),
                  )
                }
                renderItem={(item, _index, onChange) => (
                  <>
                    <div className="form-group">
                      <label>Name</label>
                      <input
                        value={item.name}
                        onChange={(e) =>
                          onChange({ ...item, name: e.target.value })
                        }
                      />
                    </div>
                    <div className="form-group">
                      <label>Link</label>
                      <input
                        value={item.link}
                        onChange={(e) =>
                          onChange({ ...item, link: e.target.value })
                        }
                      />
                    </div>
                    <div className="form-group">
                      <label>Technologies (comma-separated)</label>
                      <input
                        value={joinCsv(item.technologies)}
                        onChange={(e) =>
                          onChange({
                            ...item,
                            technologies: splitCsv(e.target.value),
                          })
                        }
                      />
                    </div>
                    <div className="form-group">
                      <label>Bullets (one per line)</label>
                      <textarea
                        value={joinMultiline(item.bullets)}
                        onChange={(e) =>
                          onChange({
                            ...item,
                            bullets: splitMultiline(e.target.value),
                          })
                        }
                      />
                    </div>
                  </>
                )}
              />
            </SectionShell>

            <SectionShell title="Skills">
              <ItemListEditor<SkillCategory>
                items={resume.skills}
                addLabel="Add Skill Category"
                onAdd={() =>
                  patchList<SkillCategory>("skills", (items) => [
                    ...items,
                    { title: "", items: [] },
                  ])
                }
                onRemove={(index) =>
                  patchList<SkillCategory>("skills", (items) =>
                    items.filter((_, i) => i !== index),
                  )
                }
                onChangeItem={(index, value) =>
                  patchList<SkillCategory>("skills", (items) =>
                    items.map((item, i) => (i === index ? value : item)),
                  )
                }
                renderItem={(item, _index, onChange) => (
                  <>
                    <div className="form-group">
                      <label>Category</label>
                      <input
                        value={item.title}
                        onChange={(e) =>
                          onChange({ ...item, title: e.target.value })
                        }
                      />
                    </div>
                    <div className="form-group">
                      <label>Items (comma-separated)</label>
                      <input
                        value={joinCsv(item.items)}
                        onChange={(e) =>
                          onChange({ ...item, items: splitCsv(e.target.value) })
                        }
                      />
                    </div>
                  </>
                )}
              />
            </SectionShell>

            <SectionShell title="Certifications">
              <ItemListEditor<Certification>
                items={resume.certifications}
                addLabel="Add Certification"
                onAdd={() =>
                  patchList<Certification>("certifications", (items) => [
                    ...items,
                    { name: "" },
                  ])
                }
                onRemove={(index) =>
                  patchList<Certification>("certifications", (items) =>
                    items.filter((_, i) => i !== index),
                  )
                }
                onChangeItem={(index, value) =>
                  patchList<Certification>("certifications", (items) =>
                    items.map((item, i) => (i === index ? value : item)),
                  )
                }
                renderItem={(item, _index, onChange) => (
                  <div className="form-group">
                    <label>Name</label>
                    <input
                      value={item.name}
                      onChange={(e) => onChange({ name: e.target.value })}
                    />
                  </div>
                )}
              />
            </SectionShell>
          </div>
        </aside>
      </div>

      {toast ? <Toast tone={toast.tone} message={toast.message} /> : null}
    </main>
  );
}
