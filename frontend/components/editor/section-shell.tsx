import { ReactNode } from "react";

type Props = {
  title: string;
  children: ReactNode;
  actions?: ReactNode;
};

export function SectionShell({ title, children, actions }: Props) {
  return (
    <section className="section">
      <h2 className="section-title">{title}</h2>
      {children}
      {actions ? <div className="section-actions">{actions}</div> : null}
    </section>
  );
}
