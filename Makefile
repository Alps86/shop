.PHONY: check static build cucumber-and-coverage race quality-gate selftest
check: static build cucumber-and-coverage race quality-gate
static build cucumber-and-coverage race quality-gate selftest:
	python3 scripts/quality.py $@
