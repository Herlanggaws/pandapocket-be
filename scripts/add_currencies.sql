-- Ensure default currencies exist (insert missing by code), including IDR.
-- Safe to re-run on databases that already have a partial default set.

INSERT INTO currencies (code, name, symbol, is_default, created_at, updated_at)
SELECT v.code, v.name, v.symbol, true, NOW(), NOW()
FROM (VALUES
    ('IDR', 'Indonesian Rupiah', 'Rp'),
    ('USD', 'US Dollar', '$'),
    ('EUR', 'Euro', '€'),
    ('GBP', 'British Pound', '£'),
    ('JPY', 'Japanese Yen', '¥'),
    ('CAD', 'Canadian Dollar', 'C$'),
    ('AUD', 'Australian Dollar', 'A$'),
    ('CHF', 'Swiss Franc', 'CHF'),
    ('CNY', 'Chinese Yuan', '¥'),
    ('INR', 'Indian Rupee', '₹'),
    ('BRL', 'Brazilian Real', 'R$'),
    ('KRW', 'South Korean Won', '₩'),
    ('MXN', 'Mexican Peso', '$'),
    ('SGD', 'Singapore Dollar', 'S$'),
    ('HKD', 'Hong Kong Dollar', 'HK$'),
    ('NZD', 'New Zealand Dollar', 'NZ$'),
    ('SEK', 'Swedish Krona', 'kr'),
    ('NOK', 'Norwegian Krone', 'kr'),
    ('DKK', 'Danish Krone', 'kr'),
    ('PLN', 'Polish Złoty', 'zł'),
    ('THB', 'Thai Baht', '฿')
) AS v(code, name, symbol)
WHERE NOT EXISTS (
    SELECT 1 FROM currencies c WHERE c.code = v.code AND c.user_id IS NULL
);

-- Verify
SELECT
    COUNT(*) AS total_currencies,
    STRING_AGG(code, ', ' ORDER BY code) AS currency_codes
FROM currencies
WHERE is_default = true AND user_id IS NULL;
