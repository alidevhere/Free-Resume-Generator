import { formatDateTime } from "@/lib/format";
import { ResumeListItem } from "@/lib/types";
import { Button } from "@/components/ui/button";

type Props = {
  item: ResumeListItem;
  onOpen: (id: number) => void;
  onDelete: (id: number, name: string) => void;
};

export function ResumeCard({ item, onOpen, onDelete }: Props) {
  return (
    <div
      className="resume-card"
      onClick={() => onOpen(item.id)}
      role="button"
      tabIndex={0}
    >
      <h3>{item.name || "Untitled Resume"}</h3>
      <p>Open this resume to edit, save, and rebuild the PDF.</p>
      <div className="resume-date">
        Created {formatDateTime(item.createdAt)}
      </div>
      <div className="resume-card-actions">
        <Button
          type="button"
          variant="danger"
          className="btn-delete"
          onClick={(event) => {
            event.stopPropagation();
            onDelete(item.id, item.name || "Untitled Resume");
          }}
        >
          Delete
        </Button>
      </div>
    </div>
  );
}
