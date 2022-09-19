.PHONY:	start
start:
	@docker compose -f stack.yml down -v
	@docker compose -f stack.yml up --build

.PHONY:	stop
stop:
	@docker compose -f stack.yml down -v
