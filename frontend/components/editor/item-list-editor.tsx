import { ReactNode } from "react";

type Props<T> = {
  items: T[];
  onAdd: () => void;
  onRemove: (index: number) => void;
  addLabel: string;
  renderItem: (
    item: T,
    index: number,
    onChange: (next: T) => void,
  ) => ReactNode;
  onChangeItem: (index: number, value: T) => void;
};

export function ItemListEditor<T>({
  items,
  onAdd,
  onRemove,
  addLabel,
  renderItem,
  onChangeItem,
}: Props<T>) {
  return (
    <>
      {items.map((item, index) => (
        <div className="item-card" key={index}>
          <button
            type="button"
            className="remove-btn"
            onClick={() => onRemove(index)}
          >
            x
          </button>
          {renderItem(item, index, (next) => onChangeItem(index, next))}
        </div>
      ))}
      <button type="button" className="btn btn-secondary" onClick={onAdd}>
        {addLabel}
      </button>
    </>
  );
}
