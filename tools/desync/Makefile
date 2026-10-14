CXX ?= clang++
CXXFLAGS ?= -std=c++17 -O2 -Wall
OPENBW ?= ../../ref/openbw
OUT ?= tracer

$(OUT): tracer.cpp
	$(CXX) $(CXXFLAGS) -isystem $(OPENBW) -o $@ tracer.cpp

clean:
	rm -f tracer

.PHONY: clean
