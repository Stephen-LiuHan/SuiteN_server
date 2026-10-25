-- Widen captures.emotion: up to 5 comma-separated stamps (e.g. "💡,🤔,⚠️,🔥,❤️" is 11 code points)
ALTER TABLE captures ALTER COLUMN emotion TYPE VARCHAR(32);
