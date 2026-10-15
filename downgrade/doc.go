// Package downgrade turns an SC:R replay into a legacy (1.16) replay in memory, so that the OpenBW
// engine can load it. It reads the replay sections itself (zlib and PKWARE compression, the 1.18-1.20
// and 1.21+ layouts), translates the commands and keeps an index from each source command to the
// engine's action stream. A legacy replay passes through unchanged.
package downgrade
