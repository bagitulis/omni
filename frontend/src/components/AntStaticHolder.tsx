/**
 * Global static holder for Ant Design v5 message / notification.
 *
 * The official antd v5 pattern: mount this inside `<AntApp>`, it
 * captures the context-aware APIs and re-exports them as module-level
 * singletons. All hooks and utilities should import `message` from
 * this file instead of from 'antd'.
 *
 * @see https://ant.design/components/app#global-scene-reduxGlobal-scene
 */
import { App } from "antd";
import { setAntStaticApi } from "./AntStaticApi";

/**
 * Render inside `<AntApp>` — captures context-aware APIs once.
 */
export function AntStaticHolder() {
  const staticFunctions = App.useApp();
  setAntStaticApi({
    message: staticFunctions.message,
    notification: staticFunctions.notification,
  });
  return null;
}
