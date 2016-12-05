package faye

const SubjectTopicsQuery = `
SELECT subject.topic_name
FROM subject
  JOIN university ON subject.university_id = university.id
WHERE university.topic_name = :topic_name
      AND subject.season = :season
      AND subject.year = :year ORDER BY subject.id
`
const DeleteSubjectQuery = `
DELETE FROM subject WHERE topic_name = :topic_name
`
