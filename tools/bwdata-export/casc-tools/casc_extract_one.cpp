// Minimal CascLib CLI: extract one file from a local CASC storage by full path.
// Usage: casc_extract_one <casc-root> <file-path-inside-casc> <output.bin>
#define __CASCLIB_SELF__
#include "CascLib.h"

#include <cstdio>
#include <cstdlib>

int main(int argc, char** argv) {
	if (argc != 4) {
		std::fprintf(stderr, "usage: %s <casc-root> <in-archive-path> <out-file>\n", argv[0]);
		return 1;
	}

	HANDLE storage = nullptr;
	if (!CascOpenStorage(argv[1], CASC_LOCALE_ALL, &storage)) {
		std::fprintf(stderr, "CascOpenStorage failed for %s\n", argv[1]);
		return 1;
	}

	HANDLE file = nullptr;
	if (!CascOpenFile(storage, argv[2], CASC_LOCALE_ALL, CASC_OPEN_BY_NAME, &file)) {
		std::fprintf(stderr, "CascOpenFile failed: %s\n", argv[2]);
		CascCloseStorage(storage);
		return 1;
	}

	FILE* out = std::fopen(argv[3], "wb");
	if (!out) {
		std::perror("fopen out");
		CascCloseFile(file);
		CascCloseStorage(storage);
		return 1;
	}

	unsigned char buf[1024 * 512];
	DWORD nread = 0;
	for (;;) {
		if (!CascReadFile(file, buf, sizeof(buf), &nread)) {
			break;
		}
		if (nread == 0) {
			break;
		}
		if (std::fwrite(buf, 1, nread, out) != nread) {
			std::perror("fwrite");
			std::fclose(out);
			CascCloseFile(file);
			CascCloseStorage(storage);
			return 1;
		}
	}

	std::fclose(out);
	CascCloseFile(file);
	CascCloseStorage(storage);
	std::fprintf(stderr, "wrote %s\n", argv[3]);
	return 0;
}
