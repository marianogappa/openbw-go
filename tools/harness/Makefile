CXX ?= clang++
CXXFLAGS ?= -std=c++17 -O2 -Wall
OPENBW ?= ../../ref/openbw

harness: main.cpp
	$(CXX) $(CXXFLAGS) -isystem $(OPENBW) -o $@ main.cpp

clean:
	rm -f harness

.PHONY: clean
