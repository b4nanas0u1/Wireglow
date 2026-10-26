package persistence

// Файл синхронизирован перед заменой; Windows не поддерживает fsync каталога через os.File.
func syncDirectory(string) error { return nil }
