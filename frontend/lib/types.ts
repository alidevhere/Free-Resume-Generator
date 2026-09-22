export type ApiEnvelope<T> = {
  success: boolean;
  message?: string;
  data: T;
};

export type Experience = {
  company: string;
  title: string;
  location: string;
  start: string;
  end: string;
  bullets: string[];
};

export type Education = {
  school: string;
  degree: string;
  location: string;
  start: string;
  end: string;
};

export type Project = {
  name: string;
  link: string;
  technologies: string[];
  bullets: string[];
};

export type OpenSourceContribution = {
  name: string;
  description: string;
  role: string;
  contribution: string;
  link: string;
  repoLink: string;
  stars: string;
  keywords: string[];
};

export type SkillCategory = {
  title: string;
  items: string[];
};

export type Certification = {
  name: string;
};

export type Resume = {
  version?: number;
  name: string;
  title?: string;
  phone: string;
  email: string;
  website: string;
  linkedin: string;
  github: string;
  template: string;
  summary: string;
  sectionOrder: string[];
  skills: SkillCategory[];
  experience: Experience[];
  projects: Project[];
  education: Education[];
  openSourceContributions: OpenSourceContribution[];
  certifications: Certification[];
};

export type ResumeListItem = {
  id: number;
  name: string;
  createdAt: string;
  updatedAt: string;
};

export type ResumeRecord = {
  id: number;
  name: string;
  createdAt: string;
  updatedAt: string;
  resume: Resume;
};

export type ResumeSavePayload = {
  name: string;
  resume: Resume;
};
