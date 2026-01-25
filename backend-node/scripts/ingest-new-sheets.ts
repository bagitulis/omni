import { google } from "googleapis";
import path from "path";

const KEY_FILE_PATH = path.join(
  __dirname,
  "../config/static/google/bertigahemat-f1bd6932b229.json"
);
const SCOPES = ["https://www.googleapis.com/auth/spreadsheets"];
const SPREADSHEET_ID = "1ZYonq5Lla0FriY-wfvqW5XEwVyfF6osB1yAtTVABFgA";

const SHEET_MAIN = "Main"; // User Database
const SHEET_CODES = "Recovery_Codes"; // 2FA/Backup Codes Database

async function main() {
  const auth = new google.auth.GoogleAuth({
    keyFile: KEY_FILE_PATH,
    scopes: SCOPES,
  });
  const sheets = google.sheets({ version: "v4", auth });

  console.log(
    "--- AUTO-IMPORT: New Sheets (@gmail.com) -> Recovery_Codes & Main ---"
  );

  try {
    // 1. Find New Sheets
    const meta = await sheets.spreadsheets.get({
      spreadsheetId: SPREADSHEET_ID,
    });
    const allSheets = meta.data.sheets || [];
    const newSheetCandidates = allSheets.filter(
      (s) => s.properties?.title?.includes("@gmail.com")
    );

    // Check if "Recovery_Codes" exists, if not check for "Vouchers" and rename, or create new
    let targetSheetTitle = SHEET_CODES;
    const voucherSheet = allSheets.find(
      (s) => s.properties?.title === "Vouchers"
    );

    if (voucherSheet) {
      console.log('Renaming "Vouchers" sheet to "Recovery_Codes"...');
      await sheets.spreadsheets.batchUpdate({
        spreadsheetId: SPREADSHEET_ID,
        requestBody: {
          requests: [
            {
              updateSheetProperties: {
                properties: {
                  sheetId: voucherSheet.properties?.sheetId,
                  title: SHEET_CODES,
                },
                fields: "title",
              },
            },
          ],
        },
      });
    }

    if (newSheetCandidates.length === 0) {
      console.log('No new sheets found (checked for "@gmail.com").');
      console.log(
        'To add data: Create a sheet named "email@gmail.com" and put codes in Column A.'
      );
      return;
    }

    console.log(`Found ${newSheetCandidates.length} new sheets to process.`);

    // 2. Fetch Existing Users from Main (to avoid duplicates)
    // Assuming Main Data starts at Row 3 (Row 1=Quota Info, Row 2=Headers)
    const mainRes = await sheets.spreadsheets.values.get({
      spreadsheetId: SPREADSHEET_ID,
      range: `${SHEET_MAIN}!A3:A`, // Read all emails
    });
    const existingEmails = new Set(
      (mainRes.data.values || []).map((r) => r[0])
    );

    // 3. Process Each New Sheet
    for (const sheet of newSheetCandidates) {
      const email = sheet.properties?.title || "";
      const sheetId = sheet.properties?.sheetId!;

      console.log(`\nProcessing: [${email}]`);

      // A. Read Codes
      const codeRes = await sheets.spreadsheets.values.get({
        spreadsheetId: SPREADSHEET_ID,
        range: `${email}!A1:A`, // Read Column A
      });

      const codes = (codeRes.data.values || [])
        .map((r) => r[0])
        .filter((c) => c && c.trim() !== ""); // Clean empty cells

      if (codes.length === 0) {
        console.log(`  -> No codes found. Deleting empty sheet...`);
      } else {
        // B. Insert into 'Recovery_Codes' Sheet
        console.log(
          `  -> Found ${codes.length} codes. moving to "${SHEET_CODES}"...`
        );

        const voucherRows = codes.map((code) => [
          code, // Code
          email, // Owner Email
          "AVAILABLE", // Status
          new Date().toISOString().split("T")[0], // Date
        ]);

        await sheets.spreadsheets.values.append({
          spreadsheetId: SPREADSHEET_ID,
          range: `${SHEET_CODES}!A1`,
          valueInputOption: "RAW",
          requestBody: { values: voucherRows },
        });

        // C. Register New User in 'Main' (if not exists)
        if (!existingEmails.has(email)) {
          console.log(`  -> New User detected! Adding to "${SHEET_MAIN}"...`);

          // Default User Row: [Email, Password, Secret, Status, User Yumna]
          // Adjust defaults as needed
          const newUserRow = [
            email, // Col 1: Email
            "123456", // Col 2: Default Pw
            "", // Col 3: Secret/Code
            "NO QUOTA", // Col 4: Status
            "New User", // Col 5: User Yumna
          ];

          await sheets.spreadsheets.values.append({
            spreadsheetId: SPREADSHEET_ID,
            range: `${SHEET_MAIN}!A1`, // Append will find list end
            valueInputOption: "RAW",
            requestBody: { values: [newUserRow] },
          });

          existingEmails.add(email); // Prevent double add in same run
        } else {
          console.log(`  -> User already exists in "${SHEET_MAIN}".`);
        }
      }

      // D. Delete the Source Sheet
      console.log(`  -> Deleting sheet "${email}"...`);
      await sheets.spreadsheets.batchUpdate({
        spreadsheetId: SPREADSHEET_ID,
        requestBody: {
          requests: [{ deleteSheet: { sheetId: sheetId } }],
        },
      });
      console.log(`  -> Done.`);
    }

    console.log("\nAll new data has been ingested successfully!");
  } catch (error: any) {
    console.error("Error:", error.message);
  }
}

main().catch(console.error);
