ALTER TABLE songs ADD COLUMN text TEXT;
UPDATE songs SET text = 'Hey Jude, don''t make it bad\nTake a sad song and make it better\nRemember to let her into your heart\nThen you can start to make it better' 
    WHERE id = 1;
UPDATE songs SET text = 'Is this the real life?\nIs this just fantasy?\nCaught in a landslide\nNo escape from reality' 
    WHERE id = 2;
UPDATE songs SET text = 'When I find myself in times of trouble\nMother Mary comes to me\nSpeaking words of wisdom\nLet it be' 
    WHERE id = 3;
UPDATE songs SET text = 'So, so you think you can tell\nHeaven from hell?\nBlue skies from pain?\nCan you tell a green field from a cold steel rail?' 
    WHERE id = 4;
UPDATE songs SET text = 'Another one bites the dust\nAnother one bites the dust\nAnd another one gone, and another one gone\nAnother one bites the dust'
     WHERE id = 5;