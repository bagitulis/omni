import {
  Alert,
  Button,
  Form,
  Input,
  Radio,
  Select,
  Space,
  Spin,
  theme,
} from "antd";
import { useEffect, useMemo, useState } from "react";
import type { Order } from "@/types/order";
import type { ShipConfirmPayload } from "@/components/modals/OrderShipModal";
import { getShippingOptions } from "@/hooks/useOrders";
import type {
  ShopeeInfoNeeded,
  ShopeePickupAddress,
  ShopeeTimeSlot,
} from "@/hooks/useOrders";

interface ShopeeShipFormValues {
  mode: "pickup" | "dropoff";
  address_id?: number;
  pickup_time_id?: string;
  branch_id?: number;
  tracking_number?: string;
}

interface ShopeeShipFormProps {
  order: Order;
  loading?: boolean;
  onClose: () => void;
  onConfirm: (payload: ShipConfirmPayload) => Promise<void>;
}

interface ShopeeOptionsState {
  pickup: ShopeePickupAddress[];
  dropoff: Array<{ branch_id: number; address: string }>;
  info_needed?: ShopeeInfoNeeded;
}

const getTimeSlots = (address?: ShopeePickupAddress): ShopeeTimeSlot[] =>
  address?.time_slots || address?.time_slot_list || [];

const hasRequiredField = (
  infoNeeded: ShopeeInfoNeeded | undefined,
  mode: "pickup" | "dropoff",
  field: string,
): boolean => {
  const fields = mode === "pickup" ? infoNeeded?.pickup : infoNeeded?.dropoff;
  return Boolean(fields?.includes(field));
};

const formatShopeeDate = (value?: string | number): string => {
  if (!value) return "";

  if (typeof value === "number") {
    const timestamp = value > 1_000_000_000_000 ? value : value * 1000;
    return new Date(timestamp).toLocaleDateString("en-US", {
      weekday: "short",
      month: "short",
      day: "numeric",
    });
  }

  const parsed = Number(value);
  if (!Number.isNaN(parsed)) {
    const timestamp = parsed > 1_000_000_000_000 ? parsed : parsed * 1000;
    return new Date(timestamp).toLocaleDateString("en-US", {
      weekday: "short",
      month: "short",
      day: "numeric",
    });
  }

  return value;
};

const getTimeSlotLabel = (slot: ShopeeTimeSlot): string => {
  const text =
    slot.pickup_time ||
    slot.time_text ||
    slot.time ||
    slot.time_slot ||
    "No label";
  const dateLabel = formatShopeeDate(slot.date);
  return dateLabel ? `${dateLabel} - ${text}` : text;
};

export function ShopeeShipForm({
  order,
  loading = false,
  onClose,
  onConfirm,
}: ShopeeShipFormProps) {
  const { token } = theme.useToken();
  const [form] = Form.useForm<ShopeeShipFormValues>();
  const [options, setOptions] = useState<ShopeeOptionsState>({
    pickup: [],
    dropoff: [],
  });
  const [fetching, setFetching] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [optionsError, setOptionsError] = useState<string | null>(null);

  useEffect(() => {
    const load = async () => {
      setFetching(true);
      setOptionsError(null);
      try {
        const response = await getShippingOptions(order.order_sn);
        setOptions({
          pickup: response.pickup || [],
          dropoff: response.dropoff || [],
          info_needed: response.info_needed,
        });
      } catch (error) {
        setOptionsError(
          error instanceof Error
            ? error.message
            : "Failed to load Shopee shipping options",
        );
      } finally {
        setFetching(false);
      }
    };
    void load();
  }, [order.order_sn]);

  const hasPickup = options.pickup.length > 0;
  const hasDropoff = options.dropoff.length > 0;
  const showTrackingFallback = !hasPickup && !hasDropoff;
  const requiredPickupByApi = (options.info_needed?.pickup || []).length > 0;
  const requiredDropoffByApi = (options.info_needed?.dropoff || []).length > 0;
  const initialMode: "pickup" | "dropoff" = useMemo(() => {
    if (requiredPickupByApi && hasPickup) return "pickup";
    if (requiredDropoffByApi && hasDropoff) return "dropoff";
    if (hasPickup) return "pickup";
    if (hasDropoff) return "dropoff";
    return "pickup";
  }, [requiredPickupByApi, requiredDropoffByApi, hasPickup, hasDropoff]);

  useEffect(() => {
    form.setFieldValue("mode", initialMode);
  }, [form, initialMode]);

  const mode = Form.useWatch("mode", form);
  const addressId = Form.useWatch("address_id", form);
  const activeMode = (mode || initialMode) as "pickup" | "dropoff";

  const selectedAddress = useMemo(
    () => options.pickup.find((item) => item.address_id === addressId),
    [options.pickup, addressId],
  );
  const timeSlotOptions = useMemo(
    () =>
      getTimeSlots(selectedAddress)
        .map((slot) => ({
          value: slot.pickup_time_id || slot.time_slot || "",
          label: getTimeSlotLabel(slot),
        }))
        .filter((slot) => slot.value),
    [selectedAddress],
  );

  const pickupTimeRequired =
    timeSlotOptions.length > 0 &&
    (hasRequiredField(options.info_needed, "pickup", "pickup_time_id") ||
      !options.info_needed);

  useEffect(() => {
    if (
      activeMode === "pickup" &&
      hasPickup &&
      !form.getFieldValue("address_id")
    ) {
      form.setFieldValue("address_id", options.pickup[0].address_id);
    }
  }, [activeMode, hasPickup, form, options.pickup]);

  useEffect(() => {
    const currentPickupTime = form.getFieldValue("pickup_time_id");
    if (
      activeMode === "pickup" &&
      pickupTimeRequired &&
      timeSlotOptions.length > 0 &&
      !currentPickupTime
    ) {
      form.setFieldValue("pickup_time_id", timeSlotOptions[0].value);
    }
  }, [activeMode, form, pickupTimeRequired, timeSlotOptions]);

  const handleSubmit = async () => {
    const values = await form.validateFields();
    const payload: ShipConfirmPayload = {
      platform: "shopee",
      data: { order_sn: order.order_sn },
    };
    if (showTrackingFallback) {
      payload.data.tracking_number = values.tracking_number?.trim();
    } else if ((values.mode || initialMode) === "pickup") {
      if (!values.address_id) {
        throw new Error("shopee API error: pickup address_id is required");
      }
      payload.data.pickup = {
        address_id: values.address_id,
        pickup_time_id: values.pickup_time_id,
      };
    } else {
      if (!values.branch_id) {
        throw new Error("shopee API error: dropoff branch_id is required");
      }
      payload.data.dropoff = { branch_id: values.branch_id };
    }

    setSubmitting(true);
    try {
      await onConfirm(payload);
      onClose();
    } finally {
      setSubmitting(false);
    }
  };

  if (fetching) return <Spin />;

  return (
    <Form form={form} layout="vertical" initialValues={{ mode: initialMode }}>
      <Alert
        type="info"
        showIcon
        message="Shopee shipment parameters"
        description="Shopee options are loaded per order and may include specific required fields."
        style={{ marginBottom: 16, borderColor: token.colorInfoBorder }}
      />

      {optionsError && (
        <Alert
          type="error"
          showIcon
          message="Failed to load Shopee shipping options"
          description={optionsError}
          style={{ marginBottom: 16, borderColor: token.colorErrorBorder }}
        />
      )}

      {hasPickup && hasDropoff && (
        <Form.Item
          name="mode"
          label="Handover Method"
          rules={[{ required: true }]}
        >
          <Radio.Group
            options={[
              { label: "Pickup", value: "pickup" },
              { label: "Dropoff", value: "dropoff" },
            ]}
          />
        </Form.Item>
      )}

      {(activeMode === "pickup" || (hasPickup && !hasDropoff)) && hasPickup && (
        <>
          <Form.Item
            name="address_id"
            label="Pickup Address"
            rules={[{ required: true, message: "Select pickup address" }]}
          >
            <Select
              options={options.pickup.map((address) => ({
                value: address.address_id,
                label: address.address,
              }))}
              placeholder="Select address"
            />
          </Form.Item>
          {timeSlotOptions.length > 0 && (
            <Form.Item
              name="pickup_time_id"
              label="Pickup Schedule"
              rules={
                pickupTimeRequired
                  ? [{ required: true, message: "Select pickup schedule" }]
                  : []
              }
            >
              <Select
                options={timeSlotOptions}
                placeholder="Select pickup timeslot"
              />
            </Form.Item>
          )}
        </>
      )}

      {(activeMode === "dropoff" || (!hasPickup && hasDropoff)) &&
        hasDropoff && (
          <Form.Item
            name="branch_id"
            label="Dropoff Branch"
            rules={[{ required: true, message: "Select dropoff branch" }]}
          >
            <Select
              options={options.dropoff.map((branch) => ({
                value: branch.branch_id,
                label: branch.address,
              }))}
              placeholder="Select dropoff branch"
            />
          </Form.Item>
        )}

      {showTrackingFallback && (
        <Form.Item
          name="tracking_number"
          label="Tracking Number"
          rules={[
            { required: true, message: "Enter tracking number" },
            { min: 5, message: "Tracking number is too short" },
          ]}
        >
          <Input placeholder="Enter tracking number" />
        </Form.Item>
      )}

      <Space style={{ justifyContent: "flex-end", width: "100%" }}>
        <Button onClick={onClose} disabled={loading || submitting}>
          Cancel
        </Button>
        <Button
          type="primary"
          loading={loading || submitting}
          onClick={() => void handleSubmit()}
        >
          Confirm Shipment
        </Button>
      </Space>
    </Form>
  );
}
