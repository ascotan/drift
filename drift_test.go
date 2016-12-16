package drift
//
// import (
//   "testing"
//   "os"
//   "strings"
//   "errors"
//   "fmt"
//   "time"
//   "path/filepath"
//   "log"
// )
//
// // ----------------------------------------------------------------------------
// // filesystem mock
// // ----------------------------------------------------------------------------
// // mock file
// type mockFile struct {
//   path  string
//   data  *strings.Reader
//   info  *mockFileInfo
// }
// func newMockFile(data string, path string, mode os.FileMode) *mockFile {
//     return &mockFile{
//       path,
//       strings.NewReader(data),
//       &mockFileInfo {
//         name:    filepath.Base(path),
//         size:    int64(len([]byte(data))),
//         mode:    mode,
//         modtime: time.Now(),
//         isdir:   false,
//         sys:     nil,
//       },
//     }
// }
// func (m *mockFile) Path() (string) { return m.path }
// func (m *mockFile) Close() error { return nil }
// func (m *mockFile) Read(p []byte) (n int, err error) { return m.data.Read(p) }
// func (m *mockFile) ReadAt(p []byte, off int64) (n int, err error) {
//   return m.data.ReadAt(p, off)
// }
// func (m *mockFile) Seek(offset int64, whence int) (int64, error) {
//   return m.data.Seek(offset, whence)
// }
// func (m *mockFile) Stat() (os.FileInfo, error) { return m.info, nil }
//
// // mock file properties
// type mockFileInfo struct {
//   name    string
//   size    int64
//   mode    os.FileMode
//   modtime time.Time
//   isdir   bool
//   sys     interface{}
// }
// func (m *mockFileInfo) Name() string { return m.name }
// func (m *mockFileInfo) Size() int64{ return m.size }
// func (m *mockFileInfo) Mode() os.FileMode { return m.mode }
// func (m *mockFileInfo) ModTime() time.Time { return m.modtime }
// func (m *mockFileInfo) IsDir() bool { return m.isdir }
// func (m *mockFileInfo) Sys() interface{} { return m.sys }
//
// // mock filesystem
// type mockFS struct{ files  map[string]file }
// func newMockFS(files ...*mockFile) *mockFS {
//   m := make(map[string]file)
//   for _, f := range files {
//     m[f.Path()] = f
//   }
//   return &mockFS{m}
// }
// func (m *mockFS) Open(name string) (file, error) {
//   val, exists := m.files[name]
//   if !exists {
//     return nil, errors.New(fmt.Sprintf("%s: no such file or directory", name))
//   }
//   return val, nil
// }
// func (m *mockFS) Stat(name string) (os.FileInfo, error) {
//   val, exists := m.files[name]
//   if !exists {
//     return nil, errors.New(fmt.Sprintf("%s: unable to stat file", name))
//   }
//   info, err := val.Stat()
//   if err != nil {
//     return nil, err
//   }
//   return info, nil
// }
// // ---------------------------------------------------------------------------
// // tests
// // ----------------------------------------------------------------------------
//
// func TestReadRevision(t *testing.T) {
//   data := `
//   -- +changeset id:hello kitty author:jgilbert dbms:ql runalways:true, runonchange:true, failonerror:true
//   -- +preconditions dbms:ql tableexists:tablename colexists:colname fkexists:fkname indexexists:indexname
//   -- +precondition-sql-check expectedResult:0 select count(*) from mytable
//   -- +precondition-sql-check expectedResult:0 select count(*) from mytable
//   -- +precondition-sql-check expectedResult:0 select count(*) from mytable
//   -- +rollback DROP TABLE xxx;
//     CREATE TABLE xxx;`
//
//   fs := newMockFS(newMockFile(data, "/tmp/migration.sql", 0644))
//   revision, err := ReadRevision("/tmp/migration.sql", fs)
//   if err != nil {
//     t.Error(err)
//   }
//   if string(revision.data) != data {
//     t.Error("Data returned is different than expected")
//   }
//
//   revision, err = ReadRevision("/tmp/does/not/exist", fs)
//   if err == nil {
//     t.Error("File should not have been found")
//   }
// }
//
// func TestParseChangesets(t *testing.T) {
//   data := `
//   -- this is a comment
//   /* this is also a comment */
//   --+ changeset id:hello kitty author:jgilbert dbms:ql runalways:true, runonchange:true, failonerror:true
//   --+ preconditions dbms:ql tableexists:tablename colexists:colname fkexists:fkname indexexists:indexname
//   --+ precondition-sql-check expectedResult:0 select count(*) from mytable
//   --+ rollback DROP TABLE xxx;
//   CREATE TABLE говорю ;
//   SELECT e.ID, e.говорю, e.DepartmentID, d.DepartmentID
//   FROM
//   	(SELECT id() AS ID, LastName, DepartmentID FROM employee) AS e,
//   	department as d,
//   WHERE e.DepartmentID == d.DepartmentID;
//   // Will work.
//
//   /* here's a multiline comment
//      that spans multiple lines */
//   --- +changeset id:2
//   CREATE TABLE exercise_logs
//       (id INTEGER PRIMARY KEY AUTOINCREMENT,
//       type TEXT,            -- 中国话不用彁字。
//       minutes INTEGER,      -- this is a comment too
//       calories INTEGER,     -- Αυτου οι θανατον μητσομαι
//       heart_rate INTEGER);  -- this is a comment too
//
//   --- +changeset id:3
//   SELECT id(), e.LastName, e.DepartmentID, d.DepartmentID
//   FROM
//   	employee AS e,
//   	department AS d,
//   WHERE e.DepartmentID == d.DepartmentID;
//   // Will always return NULL in first field.
//
//   --+ changeset id:yes
//   SELECT
//   	__Column.TableName, __Column.Ordinal, __Column.Name, __Column.Type,
//   	__Column2.NotNull, __Column2.ConstraintExpr, __Column2.DefaultExpr,
//   FROM __Column
//   LEFT JOIN __Column2 -- Αυτου οι θανατον μητσομαι
//   ON __Column.TableName == __Column2.TableName && __Column.Name == __Column2.Name
//   ORDER BY __Column.TableName, __Column.Ordinal;
//
//   --+ changeset id:no
//   BEGIN TRANSACTION
//   	UPDATE department
//   		DepartmentName = DepartmentName + " dpt.",
//   		DepartmentID = 1000+DepartmentID, -- Αυτου οι θανατον μητσομαι
//   	WHERE DepartmentID < 1000;
//   COMMIT;
//
//   -- +changeset id:hey
//   BEGIN TRANSACTION;
//   	INSERT INTO department (DepartmentID) VALUES (42);
//
//   	INSERT INTO department (
//   		DepartmentName,
//   		DepartmentID,
//   	)
//   	VALUES (
//   		"R&D",
//   		42,
//   	);
//
//   	INSERT INTO department VALUES
//   		(42, "R&D"),
//   		(17, "Sales"),
//   	;
//   COMMIT;
//
//
//   -- +changeset id:an
//   BEGIN TRANSACTION;
//   	CREATE TABLE t (
//   		a int,
//   		b int b > a && b < c DEFAULT (a+c)/2,
//   		c int,
//   );
//   COMMIT;
//   -- +changeset id:ss from
//   BEGIN TRANSACTION;
//   	CREATE TABLE department (
//   		DepartmentID   int,
//   		DepartmentName string DepartmentName IN ("HQ", "R/D", "Lab", "HR") DEFAULT "HQ",
//   	);
//   COMMIT;
//
//   -- +changeset id:fds index
//   BEGIN TRANSACTION;
//   	CREATE TABLE t (
//   		TimeStamp time TimeStamp < now() && since(TimeStamp) < duration("10s"),
//   		Event string Event != "" && Event like "[0-9]+:[ \t]+.*",
//   	);
//   COMMIT;
//   -- sql comment
//   // single line comment
//   /* here's a multiline comment
//      that spans multiple lines */
//   `
//
//   fs := newMockFS(newMockFile(data, "/tmp/migration.sql", 0644))
//   revision, _ := ReadRevision("/tmp/migration.sql", fs)
//   changesets, err := ParseChangesets(revision)
//   if err != nil {
//     t.Error(err)
//   }
//   log.Println(changesets)
// }
