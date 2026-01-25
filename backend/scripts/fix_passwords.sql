-- Update password for yumna in yumna_bertigamart tenant
UPDATE tenant_yumna_bertigamart.users 
SET password = '$2a$10$p/lnD02x6QwxnYwFuCGm0uY1orPeFbw1mEZtFA6n7y2IIKwS2J7DK', 
    failed_login_attempts = 0 
WHERE username = 'yumna';

-- Update password for yumna in tika_nusseyba tenant  
UPDATE tenant_tika_nusseyba.users 
SET password = '$2a$10$p/lnD02x6QwxnYwFuCGm0uY1orPeFbw1mEZtFA6n7y2IIKwS2J7DK', 
    failed_login_attempts = 0 
WHERE username = 'yumna';

-- Update password for tester in yumna_bertigamart tenant
UPDATE tenant_yumna_bertigamart.users 
SET password = '$2a$10$dkbCusDX2IWIF4tTfyNpFOfURoNeeuQ7sPKdDHlaa1o5rx8bgj9Q6', 
    failed_login_attempts = 0 
WHERE username = 'tester';

-- Update password for tester in tika_nusseyba tenant
UPDATE tenant_tika_nusseyba.users 
SET password = '$2a$10$dkbCusDX2IWIF4tTfyNpFOfURoNeeuQ7sPKdDHlaa1o5rx8bgj9Q6', 
    failed_login_attempts = 0 
WHERE username = 'tester';
