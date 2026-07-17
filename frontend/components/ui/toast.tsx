type Props = {
  message: string;
  tone: "success" | "error";
};

export function Toast({ message, tone }: Props) {
  return <div className={`toast show ${tone}`}>{message}</div>;
}
