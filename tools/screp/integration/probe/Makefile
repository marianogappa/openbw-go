CXX ?= clang++
CXXFLAGS ?= -std=c++17 -O2 -Wall
OPENBW ?= ../../../../ref/openbw

probe: probe.cpp
	$(CXX) $(CXXFLAGS) -isystem $(OPENBW) -o $@ probe.cpp

clean:
	rm -f probe

.PHONY: clean
