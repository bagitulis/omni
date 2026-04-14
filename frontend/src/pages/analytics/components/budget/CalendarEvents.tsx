import { Card, Typography, Tag, GlobalToken } from "antd";
import { CalendarOutlined } from "@ant-design/icons";
import type { CalendarEvent } from "@/api/analyticsIntelligence";

const { Text } = Typography;

interface CalendarEventsProps {
  events: CalendarEvent[];
  multiplier: number;
  periodDays: number;
  token: GlobalToken;
}

const eventTypeColors: Record<string, string> = {
  TWIN_DATE: "magenta",
  PAYDAY_PRIME: "gold",
  POST_PAYDAY: "orange",
  NATIONAL_HOLIDAY: "red",
  WEEKEND: "blue",
};

export const CalendarEvents = ({
  events,
  multiplier,
  periodDays,
  token,
}: CalendarEventsProps) => {
  const isFavorable = multiplier >= 1.0;

  return (
    <Card
      size="small"
      title={
        <span>
          <CalendarOutlined style={{ marginRight: 8 }} />
          Calendar Window ({periodDays} days)
        </span>
      }
      extra={
        <Tag color={isFavorable ? "green" : "orange"}>
          {multiplier.toFixed(2)}x seasonal
        </Tag>
      }
      style={{ borderRadius: token.borderRadius }}
    >
      {events.length === 0 ? (
        <Text type="secondary">No significant events in this period</Text>
      ) : (
        <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
          {events.slice(0, 8).map((evt, idx) => (
            <div
              key={`${evt.date}-${idx}`}
              style={{
                display: "flex",
                justifyContent: "space-between",
                alignItems: "center",
                padding: "4px 0",
                borderBottom:
                  idx < Math.min(events.length, 8) - 1
                    ? `1px solid ${token.colorBorderSecondary}`
                    : "none",
              }}
            >
              <div>
                <Text style={{ fontSize: 12 }}>{evt.date}</Text>
                <Text
                  type="secondary"
                  style={{ fontSize: 11, marginLeft: 8 }}
                >
                  {evt.description}
                </Text>
              </div>
              <Tag
                color={eventTypeColors[evt.type] || "default"}
                style={{ margin: 0, fontSize: 10 }}
              >
                +{((evt.multiplier - 1) * 100).toFixed(0)}%
              </Tag>
            </div>
          ))}
          {events.length > 8 && (
            <Text type="secondary" style={{ fontSize: 11 }}>
              +{events.length - 8} more events
            </Text>
          )}
        </div>
      )}
    </Card>
  );
};
