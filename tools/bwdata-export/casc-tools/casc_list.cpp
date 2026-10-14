// List files in a local CASC storage matching a CascLib find mask (e.g. "*933*", "HD2/anim/*").
// Usage: casc_list <casc-root> <mask> [listfile.csv-or-txt-or-empty]
#define __CASCLIB_SELF__
#include "CascLib.h"

#include <cstdio>
#include <cstring>

int main(int argc, char** argv) {
	if (argc < 3 || argc > 4) {
		std::fprintf(stderr, "usage: %s <casc-root> <mask> [listfile]\n", argv[0]);
		std::fprintf(stderr, "  examples: %s /Applications/StarCraft \"*933*\"\n", argv[0]);
		std::fprintf(stderr, "            %s /Applications/StarCraft \"HD2/anim/*\"\n", argv[0]);
		return 1;
	}

	const char* root = argv[1];
	const char* mask = argv[2];
	const char* listfile = (argc >= 4 && argv[3][0]) ? argv[3] : nullptr;

	HANDLE storage = nullptr;
	if (!CascOpenStorage(root, CASC_LOCALE_ALL, &storage)) {
		std::fprintf(stderr, "CascOpenStorage failed: %s\n", root);
		return 1;
	}

	CASC_FIND_DATA fd;
	std::memset(&fd, 0, sizeof(fd));
	HANDLE find = CascFindFirstFile(storage, mask, &fd, listfile);
	if (find == nullptr || find == INVALID_HANDLE_VALUE) {
		std::fprintf(stderr, "CascFindFirstFile failed mask=%s listfile=%s\n",
			mask, listfile ? listfile : "(null)");
		CascCloseStorage(storage);
		return 1;
	}

	int n = 0;
	do {
		std::printf("%s\t%llu\tfileDataId=%u\n", fd.szFileName, (unsigned long long)fd.FileSize, fd.dwFileDataId);
		n++;
	} while (CascFindNextFile(find, &fd));

	CascFindClose(find);
	CascCloseStorage(storage);
	std::fprintf(stderr, "listed %d entries\n", n);
	return 0;
}
