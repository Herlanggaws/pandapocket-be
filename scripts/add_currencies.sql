-- Add currencies script for PandaPocket
-- This script adds 20 default currencies including IDR

-- Check if currencies already exist
DO $$
DECLARE
    currency_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO currency_count FROM currencies WHERE is_default = true;
    
    IF currency_count > 0 THEN
        RAISE NOTICE 'Found % existing default currencies. Skipping creation.', currency_count;
        RAISE NOTICE 'To recreate currencies, delete existing ones first: DELETE FROM currencies WHERE is_default = true;';
    ELSE
        -- Insert currencies
        INSERT INTO currencies (code, name, symbol, is_default, created_at, updated_at) VALUES
        ('USD', 'US Dollar', '$', true, NOW(), NOW()),
        ('EUR', 'Euro', '€', true, NOW(), NOW()),
        ('GBP', 'British Pound', '£', true, NOW(), NOW()),
        ('JPY', 'Japanese Yen', '¥', true, NOW(), NOW()),
        ('AUD', 'Australian Dollar', 'A$', true, NOW(), NOW()),
        ('CAD', 'Canadian Dollar', 'C$', true, NOW(), NOW()),
        ('CHF', 'Swiss Franc', 'CHF', true, NOW(), NOW()),
        ('CNY', 'Chinese Yuan', '¥', true, NOW(), NOW()),
        ('SEK', 'Swedish Krona', 'kr', true, NOW(), NOW()),
        ('NOK', 'Norwegian Krone', 'kr', true, NOW(), NOW()),
        ('DKK', 'Danish Krone', 'kr', true, NOW(), NOW()),
        ('PLN', 'Polish Zloty', 'zł', true, NOW(), NOW()),
        ('CZK', 'Czech Koruna', 'Kč', true, NOW(), NOW()),
        ('HUF', 'Hungarian Forint', 'Ft', true, NOW(), NOW()),
        ('RUB', 'Russian Ruble', '₽', true, NOW(), NOW()),
        ('BRL', 'Brazilian Real', 'R$', true, NOW(), NOW()),
        ('INR', 'Indian Rupee', '₹', true, NOW(), NOW()),
        ('KRW', 'South Korean Won', '₩', true, NOW(), NOW()),
        ('SGD', 'Singapore Dollar', 'S$', true, NOW(), NOW()),
        ('IDR', 'Indonesian Rupiah', 'Rp', true, NOW(), NOW());
        
        RAISE NOTICE 'Successfully added 20 default currencies including IDR';
    END IF;
END $$;

-- Verify the insertion
SELECT 
    COUNT(*) as total_currencies,
    STRING_AGG(code, ', ' ORDER BY code) as currency_codes
FROM currencies 
WHERE is_default = true;
